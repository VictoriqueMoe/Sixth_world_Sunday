import { useMutation, useQueryClient } from "@tanstack/react-query";
import { cancelEvent, createEvent, deleteEvent, rsvpEvent, updateEvent, uploadEventCover } from "../endpoints";
import { queryKeys } from "../queryKeys";
import type { EventFormPayload, EventItem } from "../../types/api";

export function useCreateEvent() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (payload: EventFormPayload): Promise<EventItem> => createEvent(payload),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.events.all });
        },
    });
}

export function useUpdateEvent() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, payload }: { id: string; payload: EventFormPayload }): Promise<EventItem> =>
            updateEvent(id, payload),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.events.all });
        },
    });
}

export function useDeleteEvent() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => deleteEvent(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.events.all });
        },
    });
}

export function useCancelEvent() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => cancelEvent(id),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.events.all });
        },
    });
}

export function useRsvpEvent() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: ({ id, interested }: { id: string; interested: boolean }) => rsvpEvent(id, interested),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: queryKeys.events.all });
        },
    });
}

export function useUploadEventCover() {
    return useMutation({
        mutationFn: (file: File) => uploadEventCover(file),
    });
}
