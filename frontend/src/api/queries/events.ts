import { useQuery } from "@tanstack/react-query";
import { getEvents } from "../endpoints";
import { queryKeys } from "../queryKeys";

export function useEvents() {
    const query = useQuery({
        queryKey: queryKeys.events.all,
        queryFn: () => getEvents(),
    });
    return { events: query.data?.events ?? [], loading: query.isLoading, refresh: query.refetch };
}
