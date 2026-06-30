import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createMap, deleteMap, updateMap } from "../endpoints";
import { queryKeys } from "../queryKeys";
import type { MapItem, SaveMapPayload } from "../../types/api";

export function useCreateMap() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: SaveMapPayload): Promise<MapItem> => createMap(payload),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.maps.all });
        },
    });
}

export function useUpdateMap() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: SaveMapPayload }): Promise<MapItem> =>
            updateMap(id, payload),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.maps.all });
        },
    });
}

export function useDeleteMap() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => deleteMap(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.maps.all });
        },
    });
}
