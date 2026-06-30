import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
    deleteWeatherLocation,
    renameWeatherLocation,
    saveWeatherLocation,
    setDefaultWeatherLocation,
} from "../endpoints";
import { queryKeys } from "../queryKeys";
import type { SaveWeatherLocationPayload } from "../../types/api";

export function useSaveWeatherLocation() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: SaveWeatherLocationPayload) => saveWeatherLocation(payload),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.weather.locations });
        },
    });
}

export function useRenameWeatherLocation() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, label }: { id: string; label: string }) => renameWeatherLocation(id, label),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.weather.locations });
        },
    });
}

export function useDeleteWeatherLocation() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => deleteWeatherLocation(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.weather.locations });
        },
    });
}

export function useSetDefaultWeatherLocation() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => setDefaultWeatherLocation(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.weather.locations });
        },
    });
}
