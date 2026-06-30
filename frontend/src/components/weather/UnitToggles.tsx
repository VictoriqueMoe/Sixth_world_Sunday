import type { PrecipitationUnit, TemperatureUnit, WeatherUnits, WindSpeedUnit } from "../../utils/units";
import styles from "./weather.module.css";

interface ToggleProps<T extends string> {
    value: T;
    options: { value: T; label: string }[];
    onSelect: (value: T) => void;
}

function Toggle<T extends string>({ value, options, onSelect }: ToggleProps<T>) {
    return (
        <div className={styles.toggle}>
            {options.map(o => (
                <button
                    key={o.value}
                    type="button"
                    className={`${styles.toggleBtn}${value === o.value ? ` ${styles.toggleActive}` : ""}`}
                    onClick={() => onSelect(o.value)}
                >
                    {o.label}
                </button>
            ))}
        </div>
    );
}

interface UnitTogglesProps {
    units: WeatherUnits;
    onChange: (units: WeatherUnits) => void;
}

export function UnitToggles({ units, onChange }: UnitTogglesProps) {
    return (
        <div className={styles.unitToggles}>
            <Toggle<TemperatureUnit>
                value={units.temperature}
                onSelect={v => onChange({ ...units, temperature: v })}
                options={[
                    { value: "celsius", label: "°C" },
                    { value: "fahrenheit", label: "°F" },
                ]}
            />
            <Toggle<WindSpeedUnit>
                value={units.windSpeed}
                onSelect={v => onChange({ ...units, windSpeed: v })}
                options={[
                    { value: "kmh", label: "km/h" },
                    { value: "mph", label: "mph" },
                    { value: "ms", label: "m/s" },
                    { value: "kn", label: "kn" },
                ]}
            />
            <Toggle<PrecipitationUnit>
                value={units.precipitation}
                onSelect={v => onChange({ ...units, precipitation: v })}
                options={[
                    { value: "mm", label: "mm" },
                    { value: "inch", label: "in" },
                ]}
            />
        </div>
    );
}
