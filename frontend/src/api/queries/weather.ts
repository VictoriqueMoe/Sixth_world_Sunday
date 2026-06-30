import { useQuery } from "@tanstack/react-query";
import { getWeatherLocations } from "../endpoints";
import { getArchive, getForecast } from "../openMeteo";
import { queryKeys } from "../queryKeys";
import type { WeatherUnits } from "../../utils/units";

export function useWeatherLocations() {
    const query = useQuery({
        queryKey: queryKeys.weather.locations,
        queryFn: () => getWeatherLocations(),
    });
    return { locations: query.data?.locations ?? [], loading: query.isLoading };
}

function roundCoord(n: number): number {
    return Math.round(n * 100) / 100;
}

export function useForecast(lat: number | undefined, lon: number | undefined, units: WeatherUnits) {
    return useQuery({
        queryKey: [
            "weather",
            "forecast",
            lat != null ? roundCoord(lat) : null,
            lon != null ? roundCoord(lon) : null,
            units,
        ],
        queryFn: () => getForecast(lat as number, lon as number, units),
        enabled: lat != null && lon != null,
        staleTime: 10 * 60 * 1000,
    });
}

export function useWeatherHistory(
    lat: number | undefined,
    lon: number | undefined,
    startDate: string,
    endDate: string,
    units: WeatherUnits,
) {
    return useQuery({
        queryKey: [
            "weather",
            "history",
            lat != null ? roundCoord(lat) : null,
            lon != null ? roundCoord(lon) : null,
            startDate,
            endDate,
            units,
        ],
        queryFn: () => getArchive(lat as number, lon as number, startDate, endDate, units),
        enabled: lat != null && lon != null && startDate !== "" && endDate !== "",
        staleTime: 60 * 60 * 1000,
    });
}
