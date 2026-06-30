import type { WeatherUnits } from "../utils/units";

const GEOCODE_URL = "https://geocoding-api.open-meteo.com/v1/search";
const FORECAST_URL = "https://api.open-meteo.com/v1/forecast";
const ARCHIVE_URL = "https://archive-api.open-meteo.com/v1/archive";

export interface GeocodeResult {
    id: number;
    name: string;
    latitude: number;
    longitude: number;
    country?: string;
    country_code?: string;
    admin1?: string;
    timezone?: string;
}

export interface CurrentWeather {
    time: string;
    temperature: number;
    apparentTemperature: number;
    humidity: number;
    precipitation: number;
    weatherCode: number;
    windSpeed: number;
    windDirection: number;
    isDay: boolean;
}

export interface ForecastDay {
    date: string;
    weatherCode: number;
    tempMax: number;
    tempMin: number;
    precipitation: number;
    sunrise: string;
    sunset: string;
}

export interface ForecastResult {
    timezone: string;
    current: CurrentWeather;
    daily: ForecastDay[];
}

export interface HistoryDay {
    date: string;
    tempMax: number | null;
    tempMin: number | null;
    precipitation: number | null;
}

export interface HistoryResult {
    days: HistoryDay[];
}

function queryString(params: Record<string, string | number>): string {
    const search = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
        search.set(key, String(value));
    }
    return search.toString();
}

async function fetchJson<T>(url: string, signal?: AbortSignal): Promise<T> {
    const res = await fetch(url, { signal });
    if (!res.ok) {
        throw new Error(`Open-Meteo request failed (${res.status})`);
    }
    return (await res.json()) as T;
}

export async function geocodeSearch(name: string, signal?: AbortSignal): Promise<GeocodeResult[]> {
    const url = `${GEOCODE_URL}?${queryString({ name, count: 6, language: "en", format: "json" })}`;
    const data = await fetchJson<{ results?: GeocodeResult[] }>(url, signal);
    return data.results ?? [];
}

interface ForecastApiResponse {
    timezone: string;
    current: {
        time: string;
        temperature_2m: number;
        apparent_temperature: number;
        relative_humidity_2m: number;
        precipitation: number;
        weather_code: number;
        wind_speed_10m: number;
        wind_direction_10m: number;
        is_day: number;
    };
    daily: {
        time: string[];
        weather_code: number[];
        temperature_2m_max: number[];
        temperature_2m_min: number[];
        precipitation_sum: number[];
        sunrise: string[];
        sunset: string[];
    };
}

export async function getForecast(lat: number, lon: number, units: WeatherUnits): Promise<ForecastResult> {
    const url = `${FORECAST_URL}?${queryString({
        latitude: lat,
        longitude: lon,
        current:
            "temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,weather_code,wind_speed_10m,wind_direction_10m,is_day",
        daily: "weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum,sunrise,sunset",
        forecast_days: 7,
        timezone: "auto",
        temperature_unit: units.temperature,
        wind_speed_unit: units.windSpeed,
        precipitation_unit: units.precipitation,
    })}`;

    const data = await fetchJson<ForecastApiResponse>(url);

    const daily: ForecastDay[] = data.daily.time.map((date, i) => ({
        date,
        weatherCode: data.daily.weather_code[i],
        tempMax: data.daily.temperature_2m_max[i],
        tempMin: data.daily.temperature_2m_min[i],
        precipitation: data.daily.precipitation_sum[i],
        sunrise: data.daily.sunrise[i],
        sunset: data.daily.sunset[i],
    }));

    return {
        timezone: data.timezone,
        current: {
            time: data.current.time,
            temperature: data.current.temperature_2m,
            apparentTemperature: data.current.apparent_temperature,
            humidity: data.current.relative_humidity_2m,
            precipitation: data.current.precipitation,
            weatherCode: data.current.weather_code,
            windSpeed: data.current.wind_speed_10m,
            windDirection: data.current.wind_direction_10m,
            isDay: data.current.is_day === 1,
        },
        daily,
    };
}

interface ArchiveApiResponse {
    daily: {
        time: string[];
        temperature_2m_max: (number | null)[];
        temperature_2m_min: (number | null)[];
        precipitation_sum: (number | null)[];
    };
}

export async function getArchive(
    lat: number,
    lon: number,
    startDate: string,
    endDate: string,
    units: WeatherUnits,
): Promise<HistoryResult> {
    const url = `${ARCHIVE_URL}?${queryString({
        latitude: lat,
        longitude: lon,
        start_date: startDate,
        end_date: endDate,
        daily: "temperature_2m_max,temperature_2m_min,precipitation_sum",
        timezone: "auto",
        temperature_unit: units.temperature,
        precipitation_unit: units.precipitation,
    })}`;

    const data = await fetchJson<ArchiveApiResponse>(url);

    const days: HistoryDay[] = data.daily.time.map((date, i) => ({
        date,
        tempMax: data.daily.temperature_2m_max[i],
        tempMin: data.daily.temperature_2m_min[i],
        precipitation: data.daily.precipitation_sum[i],
    }));

    return { days };
}
