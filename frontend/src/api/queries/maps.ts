import { useQuery } from "@tanstack/react-query";
import { getMaps } from "../endpoints";
import { queryKeys } from "../queryKeys";

export function useMaps() {
    const query = useQuery({
        queryKey: queryKeys.maps.all,
        queryFn: () => getMaps(),
    });
    return { maps: query.data?.maps ?? [], loading: query.isLoading };
}
