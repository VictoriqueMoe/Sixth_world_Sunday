import React, { useRef, useState } from "react";
import { useNavigate, useLocation } from "react-router";
import { Turnstile, type TurnstileInstance } from "@marsidev/react-turnstile";
import { useAuth } from "../../hooks/useAuth";
import { usePageTitle } from "../../hooks/usePageTitle";
import { useSiteInfo } from "../../hooks/useSiteInfo";
import { Input } from "../../components/Input/Input";
import styles from "./LoginPage.module.css";

export function LoginPage() {
    usePageTitle("Sign In");
    const navigate = useNavigate();
    const location = useLocation();
    const { loginUser, registerUser } = useAuth();
    const siteInfo = useSiteInfo();
    const [isRegister, setIsRegister] = useState(false);
    const [username, setUsername] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [displayName, setDisplayName] = useState("");
    const [inviteCode, setInviteCode] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const [turnstileToken, setTurnstileToken] = useState("");
    const turnstileRef = useRef<TurnstileInstance>(null);

    const regType = siteInfo.registration_type as "open" | "invite" | "closed";
    const turnstileEnabled = siteInfo.turnstile_enabled;
    const turnstileSiteKey = siteInfo.turnstile_site_key;

    async function handleSubmit(e: React.SubmitEvent) {
        e.preventDefault();
        setError("");

        if (turnstileEnabled && !turnstileToken) {
            setError("Please complete the verification.");
            return;
        }

        setLoading(true);

        try {
            if (isRegister) {
                await registerUser(
                    username,
                    email,
                    password,
                    displayName || username,
                    inviteCode || undefined,
                    turnstileEnabled ? turnstileToken : undefined,
                );
            } else {
                await loginUser(username, password, turnstileEnabled ? turnstileToken : undefined);
            }
            navigate(location.state?.from?.pathname || "/", { replace: true });
        } catch (err) {
            setError(err instanceof Error ? err.message : "Something went wrong.");
            setTurnstileToken("");
            turnstileRef.current?.reset();
        } finally {
            setLoading(false);
        }
    }

    const canRegister = regType !== "closed";

    return (
        <div className={styles.page}>
            <div className={styles.card}>
                <div className={styles.brand}>
                    <div className={styles.badge}>
                        <span className={styles.badgeDot} /> NODE 6WS // SECURE
                    </div>

                    <div className={styles.wordmark}>
                        <div className={styles.wordmarkKicker}>Sixth World</div>
                        <h1 className={styles.wordmarkTitle}>
                            SUN<span>DAY</span>
                        </h1>
                        <p className={styles.tagline}>
                            Private node for the sprawl. Voice, text, streams, and the Archive — all jacked into one
                            grid.
                        </p>
                    </div>

                    <div className={styles.telemetry}>
                        <div className={styles.telemetryRow}>
                            <span>&gt; matrix_link</span>
                            <span className={styles.ok}>ESTABLISHED</span>
                        </div>
                        <div className={styles.telemetryRow}>
                            <span>&gt; encryption</span>
                            <span className={styles.info}>AES-256 // ACTIVE</span>
                        </div>
                        <div className={styles.telemetryRow}>
                            <span>&gt; ice_layer</span>
                            <span className={styles.warn}>SCANNING</span>
                        </div>
                        <div className={styles.telemetryRow}>
                            <span>
                                &gt; awaiting handshake<span className={styles.cursor}>_</span>
                            </span>
                            <span />
                        </div>
                    </div>
                </div>

                <div className={styles.form}>
                    <h2 className={styles.title}>{isRegister ? "New Identity" : "Jack In"}</h2>
                    <p className={styles.sub}>
                        {isRegister ? "register a handle to enter the node" : "authenticate to enter the node"}
                    </p>

                    {error && <div className={styles.error}>{error}</div>}

                    <form onSubmit={handleSubmit}>
                        <div className={styles.fieldStack}>
                            <div className={styles.field}>
                                <label className={styles.label}>Handle / SIN</label>
                                <Input
                                    className={styles.input}
                                    type="text"
                                    fullWidth
                                    placeholder="ghost_in_the_grid"
                                    value={username}
                                    onChange={e => setUsername(e.target.value)}
                                    autoComplete="username"
                                />
                            </div>

                            <div className={styles.field}>
                                <label className={styles.label}>Passkey</label>
                                <Input
                                    className={styles.input}
                                    type="password"
                                    fullWidth
                                    placeholder="••••••••••••"
                                    value={password}
                                    onChange={e => setPassword(e.target.value)}
                                    autoComplete={isRegister ? "new-password" : "current-password"}
                                />
                            </div>

                            {isRegister && (
                                <>
                                    <div className={styles.field}>
                                        <label className={styles.label}>Comm Address</label>
                                        <Input
                                            className={styles.input}
                                            type="email"
                                            fullWidth
                                            placeholder="you@thegrid.net"
                                            value={email}
                                            onChange={e => setEmail(e.target.value)}
                                            autoComplete="email"
                                        />
                                    </div>

                                    <div className={styles.field}>
                                        <label className={styles.label}>Street Name (optional)</label>
                                        <Input
                                            className={styles.input}
                                            type="text"
                                            fullWidth
                                            placeholder="Wraith"
                                            value={displayName}
                                            onChange={e => setDisplayName(e.target.value)}
                                        />
                                    </div>

                                    {regType === "invite" && (
                                        <div className={styles.field}>
                                            <label className={styles.label}>Invite Code</label>
                                            <Input
                                                className={styles.input}
                                                type="text"
                                                fullWidth
                                                placeholder="contact a fixer"
                                                value={inviteCode}
                                                onChange={e => setInviteCode(e.target.value)}
                                            />
                                        </div>
                                    )}
                                </>
                            )}

                            {turnstileEnabled && turnstileSiteKey && (
                                <div className={styles.turnstile}>
                                    <Turnstile
                                        ref={turnstileRef}
                                        siteKey={turnstileSiteKey}
                                        onSuccess={setTurnstileToken}
                                        onExpire={() => setTurnstileToken("")}
                                        options={{
                                            refreshExpired: "auto",
                                            theme: "dark",
                                        }}
                                    />
                                </div>
                            )}

                            <button
                                className={styles.btnJack}
                                type="submit"
                                disabled={
                                    !username ||
                                    !password ||
                                    loading ||
                                    (isRegister && !email) ||
                                    (isRegister && regType === "invite" && !inviteCode) ||
                                    (turnstileEnabled && !turnstileToken)
                                }
                            >
                                {loading ? "..." : isRegister ? "Register ▸" : "Jack In ▸"}
                            </button>
                        </div>
                    </form>

                    {canRegister ? (
                        <div className={styles.foot}>
                            <button className={styles.footBtn} type="button" onClick={() => setIsRegister(!isRegister)}>
                                {isRegister ? "Already a runner? Jack in" : "Request a SIN"}
                            </button>
                            {!isRegister && siteInfo.email_enabled && (
                                <button
                                    className={styles.footBtn}
                                    type="button"
                                    onClick={() => navigate("/forgot-password")}
                                >
                                    Lost passkey?
                                </button>
                            )}
                        </div>
                    ) : (
                        <div className={styles.foot}>
                            {!isRegister && <p className={styles.disabledNotice}>Registration is locked down.</p>}
                            {!isRegister && siteInfo.email_enabled && (
                                <button
                                    className={styles.footBtn}
                                    type="button"
                                    onClick={() => navigate("/forgot-password")}
                                >
                                    Lost passkey?
                                </button>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
