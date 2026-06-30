import { useEffect, useState } from "react";
import { useIsMobile } from "../../hooks/useIsMobile";
import { useWeatherLocations, useForecast } from "../../api/queries/weather";
import {
    useDeleteWeatherLocation,
    useSaveWeatherLocation,
    useSetDefaultWeatherLocation,
} from "../../api/mutations/weather";
import { loadUnits, saveUnits, type WeatherUnits } from "../../utils/units";
import type { WeatherLocation } from "../../types/api";
import { ChannelRail } from "../../components/layout/ChannelRail/ChannelRail";
import { Button } from "../../components/Button/Button";
import { LocationSearch, type SelectedLocation } from "../../components/weather/LocationSearch";
import { UnitToggles } from "../../components/weather/UnitToggles";
import { CurrentConditions } from "../../components/weather/CurrentConditions";
import { ForecastStrip } from "../../components/weather/ForecastStrip";
import { HistoryExplorer } from "../../components/weather/HistoryExplorer";
import shell from "./WeatherPage.module.css";
import styles from "../../components/weather/weather.module.css";

function sameSpot(a: { latitude: number; longitude: number }, b: { latitude: number; longitude: number }): boolean {
    return Math.abs(a.latitude - b.latitude) < 0.02 && Math.abs(a.longitude - b.longitude) < 0.02;
}

function toSelected(location: WeatherLocation): SelectedLocation {
    return {
        name: location.place_name,
        latitude: location.latitude,
        longitude: location.longitude,
        country: location.country,
        admin1: location.admin1,
        timezone: location.timezone,
    };
}

export function WeatherPage() {
    const isMobile = useIsMobile();
    const { locations } = useWeatherLocations();

    const [units, setUnits] = useState<WeatherUnits>(loadUnits);
    const [active, setActive] = useState<SelectedLocation | null>(null);
    const [seeded, setSeeded] = useState(false);
    const [toast, setToast] = useState("");

    const saveMutation = useSaveWeatherLocation();
    const deleteMutation = useDeleteWeatherLocation();
    const defaultMutation = useSetDefaultWeatherLocation();

    const forecast = useForecast(active?.latitude, active?.longitude, units);

    useEffect(() => {
        document.body.setAttribute("data-chat-page", "true");
        return () => {
            document.body.removeAttribute("data-chat-page");
        };
    }, []);

    if (!seeded && !active && locations.length > 0) {
        setSeeded(true);
        const def = locations.find(l => l.is_default) ?? locations[0];
        setActive(toSelected(def));
    }

    function updateUnits(next: WeatherUnits) {
        setUnits(next);
        saveUnits(next);
    }

    function showToast(message: string) {
        setToast(message);
        window.setTimeout(() => setToast(""), 2500);
    }

    const savedMatch = active ? (locations.find(l => sameSpot(l, active)) ?? null) : null;

    function handleSave() {
        if (!active || savedMatch) {
            return;
        }
        saveMutation.mutate(
            {
                label: "",
                place_name: active.name,
                country: active.country ?? "",
                admin1: active.admin1 ?? "",
                latitude: active.latitude,
                longitude: active.longitude,
                timezone: active.timezone ?? "",
            },
            {
                onSuccess: () => showToast("Location saved"),
                onError: () => showToast("Couldn't save location"),
            },
        );
    }

    return (
        <div className={shell.shell}>
            {!isMobile && <ChannelRail />}
            <div className={shell.main}>
                <div className={shell.page}>
                    <div className={shell.head}>
                        <h2 className={shell.title}>
                            <span className={shell.headGlyph}>{"☀"}</span>
                            Weather
                        </h2>
                        <UnitToggles units={units} onChange={updateUnits} />
                    </div>

                    <LocationSearch onSelect={setActive} onError={showToast} />

                    {locations.length > 0 && (
                        <div className={shell.savedChips}>
                            {locations.map(loc => {
                                const isActive = active != null && sameSpot(loc, active);
                                const name = loc.label.trim() || loc.place_name;
                                return (
                                    <div
                                        key={loc.id}
                                        className={`${shell.chip}${isActive ? ` ${shell.chipActive}` : ""}`}
                                    >
                                        <button
                                            type="button"
                                            className={shell.chipSelect}
                                            onClick={() => setActive(toSelected(loc))}
                                            title={loc.place_name}
                                        >
                                            {loc.is_default && <span className={shell.chipStar}>{"★"}</span>}
                                            {name}
                                        </button>
                                        {!loc.is_default && (
                                            <button
                                                type="button"
                                                className={shell.chipAction}
                                                title="Set as default"
                                                aria-label="Set as default"
                                                onClick={() => defaultMutation.mutate(loc.id)}
                                            >
                                                {"☆"}
                                            </button>
                                        )}
                                        <button
                                            type="button"
                                            className={shell.chipAction}
                                            title="Remove saved location"
                                            aria-label="Remove saved location"
                                            onClick={() => deleteMutation.mutate(loc.id)}
                                        >
                                            {"✕"}
                                        </button>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    {!active ? (
                        <div className={styles.empty}>Search for a place or use your location to see the weather.</div>
                    ) : (
                        <>
                            <div className={shell.activeHeader}>
                                <div>
                                    <div className={shell.activeName}>{active.name}</div>
                                    <div className={shell.activeMeta}>
                                        {[active.admin1, active.country].filter(Boolean).join(", ")}
                                    </div>
                                </div>
                                <Button
                                    variant={savedMatch ? "secondary" : "primary"}
                                    size="small"
                                    onClick={handleSave}
                                    disabled={!!savedMatch || saveMutation.isPending}
                                >
                                    {savedMatch ? "★ Saved" : "Save location"}
                                </Button>
                            </div>

                            {forecast.isLoading && <div className={styles.empty}>Loading weather…</div>}
                            {forecast.isError && (
                                <div className={styles.empty}>Couldn&apos;t load weather for this place.</div>
                            )}
                            {forecast.data && (
                                <>
                                    <CurrentConditions current={forecast.data.current} units={units} />
                                    <ForecastStrip daily={forecast.data.daily} units={units} />
                                </>
                            )}

                            <HistoryExplorer lat={active.latitude} lon={active.longitude} units={units} />
                        </>
                    )}

                    <footer className={shell.attribution}>
                        Weather data by{" "}
                        <a href="https://open-meteo.com/" target="_blank" rel="noreferrer">
                            Open-Meteo.com
                        </a>{" "}
                        (
                        <a href="https://creativecommons.org/licenses/by/4.0/" target="_blank" rel="noreferrer">
                            CC BY 4.0
                        </a>
                        )
                    </footer>
                </div>
            </div>

            {toast && <div className={shell.toast}>{toast}</div>}
        </div>
    );
}
