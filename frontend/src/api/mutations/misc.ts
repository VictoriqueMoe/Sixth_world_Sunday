import { useMutation, useQueryClient } from "@tanstack/react-query";
import { blockUser, createReport, unblockUser } from "../endpoints";

export function useBlockUser() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => blockUser(id),
        onSuccess: (_d, id) => {
            qc.invalidateQueries({ queryKey: ["block-status", id] });
            qc.invalidateQueries({ queryKey: ["blocked-users"] });
        },
    });
}

export function useUnblockUser() {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => unblockUser(id),
        onSuccess: (_d, id) => {
            qc.invalidateQueries({ queryKey: ["block-status", id] });
            qc.invalidateQueries({ queryKey: ["blocked-users"] });
        },
    });
}

export function useCreateReport() {
    return useMutation({
        mutationFn: ({
            targetType,
            targetId,
            reason,
            contextId,
        }: {
            targetType: string;
            targetId: string;
            reason: string;
            contextId?: string;
        }) => createReport(targetType, targetId, reason, contextId),
    });
}
