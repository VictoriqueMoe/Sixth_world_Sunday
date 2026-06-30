export type TemperatureUnit = "celsius" | "fahrenheit";
export type WindSpeedUnit = "kmh" | "ms" | "mph" | "kn";
export type PrecipitationUnit = "mm" | "inch";

export interface WeatherUnits {
    temperature: TemperatureUnit;
    windSpeed: WindSpeedUnit;
    precipitation: PrecipitationUnit;
}

const STORAGE_KEY = "weather-units";

function detectRegion(): string {
    try {
        const region = new Intl.Locale(navigator.language).maximize().region;
        if (region) {
            return region.toUpperCase();
        }
    } catch {
        // fall through to manual parse
    }

    const parts = navigator.language.split("-");
    return parts.length > 1 ? parts[parts.length - 1].toUpperCase() : "";
}

export function localeDefaultUnits(): WeatherUnits {
    const region = detectRegion();
    if (region === "US") {
        return { temperature: "fahrenheit", windSpeed: "mph", precipitation: "inch" };
    }
    if (region === "GB") {
        return { temperature: "celsius", windSpeed: "mph", precipitation: "mm" };
    }
    return { temperature: "celsius", windSpeed: "kmh", precipitation: "mm" };
}

export function loadUnits(): WeatherUnits {
    const defaults = localeDefaultUnits();
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw) {
            const parsed = JSON.parse(raw) as Partial<WeatherUnits>;
            return {
                temperature: parsed.temperature ?? defaults.temperature,
                windSpeed: parsed.windSpeed ?? defaults.windSpeed,
                precipitation: parsed.precipitation ?? defaults.precipitation,
            };
        }
    } catch {
        // ignore malformed storage and fall back to locale defaults
    }
    return defaults;
}

export function saveUnits(units: WeatherUnits): void {
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(units));
    } catch {
        // ignore storage write failures (private mode, quota)
    }
}

export function temperatureSuffix(unit: TemperatureUnit): string {
    return unit === "fahrenheit" ? "°F" : "°C";
}

const WIND_SUFFIX: Record<WindSpeedUnit, string> = {
    kmh: "km/h",
    ms: "m/s",
    mph: "mph",
    kn: "kn",
};

export function windSuffix(unit: WindSpeedUnit): string {
    return WIND_SUFFIX[unit];
}

export function precipitationSuffix(unit: PrecipitationUnit): string {
    return unit === "inch" ? "in" : "mm";
}
