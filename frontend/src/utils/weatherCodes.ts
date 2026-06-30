export interface WeatherCodeInfo {
    label: string;
    glyph: string;
}

// WMO weather interpretation codes. Glyphs use monochrome dingbats (not colour
// emoji) so they pick up the theme's text colour like the rest of the UI.
const WEATHER_CODES: Record<number, WeatherCodeInfo> = {
    0: { label: "Clear sky", glyph: "☀" },
    1: { label: "Mainly clear", glyph: "☀" },
    2: { label: "Partly cloudy", glyph: "☁" },
    3: { label: "Overcast", glyph: "☁" },
    45: { label: "Fog", glyph: "☁" },
    48: { label: "Rime fog", glyph: "☁" },
    51: { label: "Light drizzle", glyph: "☂" },
    53: { label: "Drizzle", glyph: "☂" },
    55: { label: "Dense drizzle", glyph: "☂" },
    56: { label: "Freezing drizzle", glyph: "☂" },
    57: { label: "Freezing drizzle", glyph: "☂" },
    61: { label: "Light rain", glyph: "☂" },
    63: { label: "Rain", glyph: "☂" },
    65: { label: "Heavy rain", glyph: "☂" },
    66: { label: "Freezing rain", glyph: "☂" },
    67: { label: "Freezing rain", glyph: "☂" },
    71: { label: "Light snow", glyph: "❄" },
    73: { label: "Snow", glyph: "❄" },
    75: { label: "Heavy snow", glyph: "❄" },
    77: { label: "Snow grains", glyph: "❄" },
    80: { label: "Rain showers", glyph: "☂" },
    81: { label: "Rain showers", glyph: "☂" },
    82: { label: "Violent rain showers", glyph: "☂" },
    85: { label: "Snow showers", glyph: "❄" },
    86: { label: "Heavy snow showers", glyph: "❄" },
    95: { label: "Thunderstorm", glyph: "☈" },
    96: { label: "Thunderstorm with hail", glyph: "☈" },
    99: { label: "Thunderstorm with hail", glyph: "☈" },
};

export function weatherCodeInfo(code: number): WeatherCodeInfo {
    return WEATHER_CODES[code] ?? { label: "Unknown", glyph: "☁" };
}
