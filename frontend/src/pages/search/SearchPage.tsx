import type { SubmitEvent } from "react";
import { useSearchParams } from "react-router";
import { useSiteSearch } from "../../api/queries/search";
import { Pagination } from "../../components/Pagination/Pagination";
import { Input } from "../../components/Input/Input";
import { Button } from "../../components/Button/Button";
import { InfoPanel } from "../../components/InfoPanel/InfoPanel";
import { SearchResultRow } from "../../components/layout/GlobalSearch/SearchResultRow";
import { SEARCH_FILTER_OPTIONS } from "../../components/layout/GlobalSearch/searchTypeMeta";
import { usePageTitle } from "../../hooks/usePageTitle";
import styles from "./SearchPage.module.css";

const PAGE_LIMIT = 20;

export function SearchPage() {
    const [params, setParams] = useSearchParams();
    const queryParam = params.get("q") ?? "";
    const typeParam = params.get("type") ?? "";
    const pageParam = Math.max(0, Number(params.get("page") ?? "0"));

    usePageTitle(queryParam ? `Search: ${queryParam}` : "Search");

    const offset = pageParam * PAGE_LIMIT;
    const { results, total, loading, fetching } = useSiteSearch(queryParam, typeParam, PAGE_LIMIT, offset);

    function setQ(next: string) {
        const trimmed = next.trim();
        const newParams = new URLSearchParams(params);
        if (trimmed) {
            newParams.set("q", trimmed);
        } else {
            newParams.delete("q");
        }
        newParams.delete("page");
        setParams(newParams);
    }

    function setType(value: string) {
        const newParams = new URLSearchParams(params);
        if (value) {
            newParams.set("type", value);
        } else {
            newParams.delete("type");
        }
        newParams.delete("page");
        setParams(newParams);
    }

    function setPage(next: number) {
        const newParams = new URLSearchParams(params);
        if (next > 0) {
            newParams.set("page", String(next));
        } else {
            newParams.delete("page");
        }
        setParams(newParams);
    }

    function handleSubmit(e: SubmitEvent<HTMLFormElement>) {
        e.preventDefault();
        const data = new FormData(e.currentTarget);
        setQ(String(data.get("q") ?? ""));
    }

    return (
        <div className={styles.page}>
            <h1 className={styles.heading}>Search the Grid</h1>

            <InfoPanel title="Query Syntax">
                <p>
                    Most queries just work — type handles, fragments, even typos and misspellings ("wrth" finds
                    "Wraith"). For when you need more precision:
                </p>
                <ul className={styles.syntaxList}>
                    <li>
                        <code>wraith decker</code> — both words must appear
                    </li>
                    <li>
                        <code>wraith OR ghost</code> — either word matches
                    </li>
                    <li>
                        <code>runner -decker</code> — has <strong>runner</strong>, but not <strong>decker</strong>
                    </li>
                    <li>
                        <code>"black ice"</code> — exact phrase, words adjacent in order
                    </li>
                    <li>
                        Mix freely: <code>"shadow run" wraith -decker</code>
                    </li>
                </ul>
                <p>
                    Use the chips below to limit results to a section. <strong>Comments only</strong> spans every
                    section. Drafts and banned-user paydata never surface in results.
                </p>
            </InfoPanel>

            <form className={styles.searchForm} onSubmit={handleSubmit}>
                <Input
                    key={queryParam}
                    name="q"
                    type="search"
                    placeholder="Search the grid..."
                    defaultValue={queryParam}
                    fullWidth
                    autoFocus
                />
                <Button type="submit" variant="primary">
                    Search
                </Button>
            </form>

            <div className={styles.filters}>
                {SEARCH_FILTER_OPTIONS.map(opt => (
                    <button
                        key={opt.value || "all"}
                        type="button"
                        className={`${styles.chip} ${typeParam === opt.value ? styles.chipActive : ""}`}
                        onClick={() => setType(opt.value)}
                    >
                        {opt.label}
                    </button>
                ))}
            </div>

            {!queryParam && <div className={styles.hint}>Enter at least 2 characters to query the grid.</div>}

            {queryParam && queryParam.trim().length < 2 && (
                <div className={styles.hint}>Enter at least 2 characters.</div>
            )}

            {queryParam && queryParam.trim().length >= 2 && (
                <>
                    <div className={styles.summary}>
                        {loading ? "Scanning..." : `${total} ${total === 1 ? "hit" : "hits"} for "${queryParam}"`}
                        {fetching && !loading && <span className={styles.refetch}> updating...</span>}
                    </div>

                    {!loading && results.length === 0 && (
                        <div className={styles.empty}>No paydata found. Try different keywords or drop the filter.</div>
                    )}

                    <div className={styles.results}>
                        {results.map(r => (
                            <SearchResultRow key={`${r.type}-${r.id}`} result={r} variant="page" />
                        ))}
                    </div>

                    <Pagination
                        offset={offset}
                        limit={PAGE_LIMIT}
                        total={total}
                        hasNext={offset + PAGE_LIMIT < total}
                        hasPrev={pageParam > 0}
                        onNext={() => setPage(pageParam + 1)}
                        onPrev={() => setPage(Math.max(0, pageParam - 1))}
                    />
                </>
            )}
        </div>
    );
}
