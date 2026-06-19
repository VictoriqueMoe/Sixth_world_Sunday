import { useQuery } from "@tanstack/react-query";
import { getBlockedUsers, getUserRooms } from "../endpoints";
import { queryKeys } from "../queryKeys";

export function useBlockedUsers(userID: string) {
    const query = useQuery({
        queryKey: queryKeys.profile.blockedUsers(userID),
        queryFn: () => getBlockedUsers(),
        enabled: !!userID,
    });
    return { blocked: query.data?.users ?? [], loading: query.isLoading, refresh: query.refetch };
}

export function useUserRooms() {
    const query = useQuery({
        queryKey: ["user", "chat-rooms"],
        queryFn: () => getUserRooms(),
    });
    return { rooms: query.data?.rooms ?? [], loading: query.isLoading };
}
