package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"Sixth_world_Sunday/internal/admin"
	"Sixth_world_Sunday/internal/auth"
	"Sixth_world_Sunday/internal/authz"
	blocksvc "Sixth_world_Sunday/internal/block"
	"Sixth_world_Sunday/internal/chat"
	"Sixth_world_Sunday/internal/config"
	"Sixth_world_Sunday/internal/contentfilter"
	"Sixth_world_Sunday/internal/controllers"
	"Sixth_world_Sunday/internal/email"
	"Sixth_world_Sunday/internal/filehost"
	"Sixth_world_Sunday/internal/logger"
	"Sixth_world_Sunday/internal/media"
	"Sixth_world_Sunday/internal/middleware"
	"Sixth_world_Sunday/internal/notification"
	"Sixth_world_Sunday/internal/og"
	"Sixth_world_Sunday/internal/profile"
	"Sixth_world_Sunday/internal/report"
	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/routes"
	searchsvc "Sixth_world_Sunday/internal/search"
	"Sixth_world_Sunday/internal/session"
	"Sixth_world_Sunday/internal/settings"
	"Sixth_world_Sunday/internal/upload"
	"Sixth_world_Sunday/internal/user"
	"Sixth_world_Sunday/internal/vanityrole"
	"Sixth_world_Sunday/internal/ws"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

var (
	//go:embed static/*
	staticFiles embed.FS
)

type (
	services struct {
		settings      settings.Service
		auth          auth.Service
		profile       profile.Service
		notification  notification.Service
		admin         admin.Service
		authz         authz.Service
		chat          chat.Service
		report        report.Service
		block         blocksvc.Service
		email         email.Service
		session       *session.Manager
		upload        upload.Service
		hub           *ws.Hub
		mediaProc     *media.Processor
		contentFilter *contentfilter.Manager
		vanityRole    vanityrole.Service
		search        searchsvc.Service
		user          user.Service
		fileVault     filehost.Service
	}
)

func initServer() *fiber.App {
	repos, settingsSvc := initDatabase()
	svc := initServices(repos, settingsSvc)
	app := initApp(svc, repos, settingsSvc)
	registerListeners(settingsSvc, app, svc, repos)
	return app
}

func initApp(svc *services, repos *repository.Repositories, settingsSvc settings.Service) *fiber.App {
	app := fiber.New(fiber.Config{
		ProxyHeader: "CF-Connecting-IP",
		TrustProxy:  true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Loopback: true,
			Private:  true,
		},
	})

	middleware.Setup(app, settingsSvc, svc.session, svc.authz)
	app.Use(middleware.Metrics())
	app.Get("/metrics", middleware.MetricsHandler())
	registerPprofRoutes(app, svc.session, svc.authz)

	lastSeenIP := middleware.NewLastSeenIP(repos.User, time.Hour)
	app.Use(middleware.RecordLastSeenIP(lastSeenIP))

	htmlBytes, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to read index.html from embedded files")
	}

	ctrlService := controllers.NewService(
		svc.auth, svc.profile, svc.notification, svc.admin,
		svc.authz, settingsSvc, svc.chat, svc.report,
		svc.block, svc.user, svc.upload, svc.mediaProc, svc.vanityRole, svc.session, svc.hub, svc.search,
		svc.fileVault,
	)
	routes.PublicRoutes(ctrlService, app)

	ogResolver := og.NewResolver(
		settingsSvc,
		string(htmlBytes),
		settingsSvc.Get(context.Background(), config.SettingBaseURL),
	)

	app.Get("/api/v1/ws", ws.Handler(svc.hub, svc.session, svc.chat, func() string {
		return settingsSvc.Get(context.Background(), config.SettingBaseURL)
	}))
	uploadsHandler := static.New(svc.upload.GetUploadDir(), static.Config{
		Browse: false,
	})
	app.Get("/uploads/*", middleware.RequireAuthGate(svc.session, svc.authz), uploadsHandler)

	ogImageHandler := controllers.NewOGImageHandler(svc.upload.GetUploadDir())
	ogImageHandler.Register(app)

	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to create static sub-filesystem")
	}

	embeddedStaticHandler := static.New("", static.Config{
		FS: staticFS,
	})

	app.Get("/*", func(ctx fiber.Ctx) error {
		path := ctx.Path()
		if strings.Contains(path, ".") {
			return embeddedStaticHandler(ctx)
		}
		return ctx.Type("html").SendString(ogResolver.Resolve(ctx.Context(), path))
	})

	logRoutes(app)

	return app
}

func logRoutes(app *fiber.App) {
	rs := app.GetRoutes(true)

	if logger.Log.Debug().Enabled() {
		sort.Slice(rs, func(i, j int) bool {
			if rs[i].Path == rs[j].Path {
				return rs[i].Method < rs[j].Method
			}
			return rs[i].Path < rs[j].Path
		})

		methodWidth := len("METHOD")
		pathWidth := len("PATH")
		for _, r := range rs {
			if len(r.Method) > methodWidth {
				methodWidth = len(r.Method)
			}
			if len(r.Path) > pathWidth {
				pathWidth = len(r.Path)
			}
		}

		border := "+" + strings.Repeat("-", methodWidth+2) + "+" + strings.Repeat("-", pathWidth+2) + "+"
		var b strings.Builder
		b.WriteString("\n")
		b.WriteString(border + "\n")
		b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", methodWidth, "METHOD", pathWidth, "PATH"))
		b.WriteString(border + "\n")
		for _, r := range rs {
			b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", methodWidth, r.Method, pathWidth, r.Path))
		}
		b.WriteString(border)

		logger.Log.Debug().Msgf("registered routes:%s", b.String())
	}

	logger.Log.Info().Msgf("%d routes mounted", len(rs))
}
