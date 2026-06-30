import { useState } from "react";
import { useWeatherHistory } from "../../api/queries/weather";
import { precipitationSuffix, temperatureSuffix, type WeatherUnits } from "../../utils/units";
import { Input } from "../Input/Input";
import { WeatherChart } from "./WeatherChart";
import styles from "./weather.module.css";

interface HistoryExplorerProps {
    lat: number;
    lon: number;
    units: WeatherUnits;
}

function isoDaysAgo(days: number): string {
    const d = new Date();
    d.setDate(d.getDate() - days);
    return d.toISOString().slice(0, 10);
}

export function HistoryExplorer({ lat, lon, units }: HistoryExplorerProps) {
    const [start, setStart] = useState(() => isoDaysAgo(35));
    const [end, setEnd] = useState(() => isoDaysAgo(5));
    const today = isoDaysAgo(0);
    const { data, isLoading, isError } = useWeatherHistory(lat, lon, start, end, units);

    return (
        <section className={styles.history}>
            <div className={styles.historyHead}>
                <h3 className={styles.sectionTitle}>Historical</h3>
                <div className={styles.historyRange}>
                    <Input type="date" value={start} max={end} onChange={e => setStart(e.target.value)} />
                    <span className={styles.rangeSep}>→</span>
                    <Input type="date" value={end} min={start} max={today} onChange={e => setEnd(e.target.value)} />
                </div>
            </div>
            <p className={styles.historyNote}>Archive data trails real time by roughly 5 days.</p>

            {isLoading && <div className={styles.chartEmpty}>Loading history…</div>}
            {isError && <div className={styles.chartEmpty}>Couldn&apos;t load historical data.</div>}

            {data && (
                <>
                    <WeatherChart days={data.days} />
                    <div className={styles.legend}>
                        <span className={styles.legendHi}>High</span>
                        <span className={styles.legendLo}>Low</span>
                    </div>
                    <div className={styles.tableWrap}>
                        <table className={styles.histTable}>
                            <thead>
                                <tr>
                                    <th>Date</th>
                                    <th>High</th>
                                    <th>Low</th>
                                    <th>Precip</th>
                                </tr>
                            </thead>
                            <tbody>
                                {data.days.map(d => (
                                    <tr key={d.date}>
                                        <td>{d.date}</td>
                                        <td>
                                            {d.tempMax != null
                                                ? `${Math.round(d.tempMax)}${temperatureSuffix(units.temperature)}`
                                                : "—"}
                                        </td>
                                        <td>
                                            {d.tempMin != null
                                                ? `${Math.round(d.tempMin)}${temperatureSuffix(units.temperature)}`
                                                : "—"}
                                        </td>
                                        <td>
                                            {d.precipitation != null
                                                ? `${d.precipitation} ${precipitationSuffix(units.precipitation)}`
                                                : "—"}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </>
            )}
        </section>
    );
}
