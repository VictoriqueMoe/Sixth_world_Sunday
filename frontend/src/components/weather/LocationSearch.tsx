import { useEffect, useState } from "react";
import { geocodeSearch, type GeocodeResult } from "../../api/openMeteo";
import { Input } from "../Input/Input";
import { Button } from "../Button/Button";
import styles from "./weather.module.css";

export interface SelectedLocation {
    name: string;
    latitude: number;
    longitude: number;
    country?: string;
    admin1?: string;
    timezone?: string;
}

interface LocationSearchProps {
    onSelect: (location: SelectedLocation) => void;
    onError: (message: string) => void;
}

export function LocationSearch({ onSelect, onError }: LocationSearchProps) {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<GeocodeResult[]>([]);
    const [open, setOpen] = useState(false);
    const [locating, setLocating] = useState(false);

    useEffect(() => {
        const q = query.trim();
        const controller = new AbortController();
        const timer = setTimeout(() => {
            if (q.length < 2) {
                setResults([]);
                setOpen(false);
                return;
            }
            geocodeSearch(q, controller.signal)
                .then(found => {
                    setResults(found);
                    setOpen(true);
                })
                .catch(() => {
                    // aborted or network error — leave previous results
                });
        }, 300);

        return () => {
            controller.abort();
            clearTimeout(timer);
        };
    }, [query]);

    function choose(result: GeocodeResult) {
        onSelect({
            name: result.name,
            latitude: result.latitude,
            longitude: result.longitude,
            country: result.country,
            admin1: result.admin1,
            timezone: result.timezone,
        });
        setQuery("");
        setResults([]);
        setOpen(false);
    }

    function useMyLocation() {
        if (!navigator.geolocation) {
            onError("Geolocation isn't available in this browser.");
            return;
        }

        setLocating(true);
        navigator.geolocation.getCurrentPosition(
            position => {
                setLocating(false);
                onSelect({
                    name: "My location",
                    latitude: position.coords.latitude,
                    longitude: position.coords.longitude,
                });
            },
            () => {
                setLocating(false);
                onError("Couldn't get your location.");
            },
            { enableHighAccuracy: false, timeout: 10000 },
        );
    }

    return (
        <div className={styles.search}>
            <div className={styles.searchInputWrap}>
                <Input
                    fullWidth
                    type="text"
                    value={query}
                    placeholder="Search a city…"
                    onChange={e => setQuery(e.target.value)}
                    onFocus={() => results.length > 0 && setOpen(true)}
                />
                {open && results.length > 0 && (
                    <ul className={styles.searchResults}>
                        {results.map(result => (
                            <li key={result.id}>
                                <button type="button" className={styles.searchResult} onClick={() => choose(result)}>
                                    <span className={styles.resultName}>{result.name}</span>
                                    <span className={styles.resultMeta}>
                                        {[result.admin1, result.country].filter(Boolean).join(", ")}
                                    </span>
                                </button>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
            <Button variant="secondary" size="small" onClick={useMyLocation} disabled={locating}>
                {locating ? "Locating…" : "⌖ Use my location"}
            </Button>
        </div>
    );
}
