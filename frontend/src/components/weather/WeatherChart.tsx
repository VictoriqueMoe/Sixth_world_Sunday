import type { HistoryDay } from "../../api/openMeteo";
import styles from "./weather.module.css";

interface WeatherChartProps {
    days: HistoryDay[];
}

const WIDTH = 800;
const HEIGHT = 220;
const PAD_X = 8;
const PAD_Y = 18;

export function WeatherChart({ days }: WeatherChartProps) {
    const valid = days.filter(d => d.tempMax != null && d.tempMin != null);
    if (valid.length === 0) {
        return <div className={styles.chartEmpty}>No data for this range.</div>;
    }

    const hi = Math.max(...valid.map(d => d.tempMax as number));
    const lo = Math.min(...valid.map(d => d.tempMin as number));
    const range = hi - lo || 1;

    const x = (i: number) => PAD_X + (i * (WIDTH - PAD_X * 2)) / Math.max(valid.length - 1, 1);
    const y = (t: number) => PAD_Y + (HEIGHT - PAD_Y * 2) * (1 - (t - lo) / range);

    const line = (pick: (d: HistoryDay) => number) =>
        valid.map((d, i) => `${i === 0 ? "M" : "L"}${x(i).toFixed(1)},${y(pick(d)).toFixed(1)}`).join(" ");

    const hiPath = line(d => d.tempMax as number);
    const loPath = line(d => d.tempMin as number);

    return (
        <svg
            className={styles.chart}
            viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
            preserveAspectRatio="none"
            role="img"
            aria-label="Daily high and low temperature history"
        >
            <text x={PAD_X} y={PAD_Y - 4} className={styles.chartTick}>
                {Math.round(hi)}°
            </text>
            <text x={PAD_X} y={HEIGHT - 4} className={styles.chartTick}>
                {Math.round(lo)}°
            </text>
            <path d={hiPath} className={styles.chartHi} />
            <path d={loPath} className={styles.chartLo} />
        </svg>
    );
}
