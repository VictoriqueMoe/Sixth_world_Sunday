import type { ForecastDay } from "../../api/openMeteo";
import { weatherCodeInfo } from "../../utils/weatherCodes";
import { precipitationSuffix, type WeatherUnits } from "../../utils/units";
import styles from "./weather.module.css";

interface ForecastStripProps {
    daily: ForecastDay[];
    units: WeatherUnits;
}

function dayLabel(date: string, index: number): string {
    if (index === 0) {
        return "Today";
    }
    return new Date(`${date}T00:00:00`).toLocaleDateString(undefined, { weekday: "short" });
}

export function ForecastStrip({ daily, units }: ForecastStripProps) {
    return (
        <section className={styles.forecast}>
            {daily.map((day, index) => {
                const info = weatherCodeInfo(day.weatherCode);
                return (
                    <div key={day.date} className={styles.forecastDay}>
                        <div className={styles.forecastName}>{dayLabel(day.date, index)}</div>
                        <div className={styles.forecastGlyph} title={info.label}>
                            {info.glyph}
                        </div>
                        <div className={styles.forecastTemps}>
                            <span className={styles.tHi}>{Math.round(day.tempMax)}°</span>
                            <span className={styles.tLo}>{Math.round(day.tempMin)}°</span>
                        </div>
                        <div className={styles.forecastPrecip}>
                            {day.precipitation} {precipitationSuffix(units.precipitation)}
                        </div>
                    </div>
                );
            })}
        </section>
    );
}
