import { useQuery } from "@tanstack/react-query";
import { getBlockStatus, getRules, listUsersPublic, searchUsers } from "../endpoints";
import { queryClient } from "../queryClient";

export function fetchSearchUsers(query: string) {
    return queryClient.fetchQuery({
        queryKey: ["users", "search", query],
        queryFn: () => searchUsers(query),
    });
}

export function useSearchUsers(query: string, enabled = true) {
    const q = useQuery({
        queryKey: ["users", "search", query],
        queryFn: () => searchUsers(query),
        enabled: enabled && !!query,
    });
    return { users: q.data ?? [], loading: q.isLoading };
}

export function useUsersPublic() {
    const q = useQuery({ queryKey: ["users", "public"], queryFn: () => listUsersPublic() });
    return { users: q.data ?? [], loading: q.isLoading };
}

export function useBlockStatus(userId: string) {
    const q = useQuery({
        queryKey: ["block-status", userId],
        queryFn: () => getBlockStatus(userId),
        enabled: !!userId,
    });
    return {
        status: q.data ?? { blocking: false, blocked_by: false },
        loading: q.isLoading,
        refresh: q.refetch,
    };
}

export function useRules(page: string) {
    const q = useQuery({
        queryKey: ["rules", page],
        queryFn: () => getRules(page),
        enabled: !!page,
    });
    return { rules: q.data?.rules ?? "", loading: q.isLoading };
}
