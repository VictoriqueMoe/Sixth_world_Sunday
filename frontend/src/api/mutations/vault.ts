import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
    createVaultFolder,
    deleteVaultFile,
    deleteVaultFolder,
    renameVaultFile,
    renameVaultFolder,
    setVaultFileLocked,
    setVaultFolderLocked,
    uploadVaultFile,
} from "../endpoints";

function useVaultInvalidate() {
    const qc = useQueryClient();
    return () => qc.invalidateQueries({ queryKey: ["vault"] });
}

export function useCreateVaultFolder() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ name, parentId }: { name: string; parentId: string | null }) =>
            createVaultFolder(name, parentId),
        onSuccess: invalidate,
    });
}

export function useRenameVaultFolder() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ id, name }: { id: string; name: string }) => renameVaultFolder(id, name),
        onSuccess: invalidate,
    });
}

export function useDeleteVaultFolder() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: (id: string) => deleteVaultFolder(id),
        onSuccess: invalidate,
    });
}

export function useSetVaultFolderLocked() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ id, locked }: { id: string; locked: boolean }) => setVaultFolderLocked(id, locked),
        onSuccess: invalidate,
    });
}

export function useUploadVaultFile() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ folderId, file }: { folderId: string | null; file: File }) => uploadVaultFile(folderId, file),
        onSuccess: invalidate,
    });
}

export function useRenameVaultFile() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ id, name }: { id: string; name: string }) => renameVaultFile(id, name),
        onSuccess: invalidate,
    });
}

export function useDeleteVaultFile() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: (id: string) => deleteVaultFile(id),
        onSuccess: invalidate,
    });
}

export function useSetVaultFileLocked() {
    const invalidate = useVaultInvalidate();
    return useMutation({
        mutationFn: ({ id, locked }: { id: string; locked: boolean }) => setVaultFileLocked(id, locked),
        onSuccess: invalidate,
    });
}
