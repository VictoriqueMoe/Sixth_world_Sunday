import { useQuery } from "@tanstack/react-query";
import { browseVault } from "../endpoints";

export function useVaultBrowse(folderId: string | null) {
    const query = useQuery({
        queryKey: ["vault", "browse", folderId ?? "root"],
        queryFn: () => browseVault(folderId),
    });
    return {
        data: query.data ?? null,
        loading: query.isLoading,
        error: query.error as Error | null,
        refresh: query.refetch,
    };
}
