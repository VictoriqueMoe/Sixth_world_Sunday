import type { CurrentWeather } from "../../api/openMeteo";
import { weatherCodeInfo } from "../../utils/weatherCodes";
import { precipitationSuffix, temperatureSuffix, windSuffix, type WeatherUnits } from "../../utils/units";
import styles from "./weather.module.css";

interface CurrentConditionsProps {
    current: CurrentWeather;
    units: WeatherUnits;
}

const CARDINALS = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"];

function windCardinal(deg: number): string {
    return CARDINALS[Math.round(deg / 45) % 8];
}

export function CurrentConditions({ current, units }: CurrentConditionsProps) {
    const info = weatherCodeInfo(current.weatherCode);
    const temp = temperatureSuffix(units.temperature);

    return (
        <section className={styles.currentCard}>
            <div className={styles.currentMain}>
                <span className={styles.currentGlyph} title={info.label}>
                    {info.glyph}
                </span>
                <div>
                    <div className={styles.currentTemp}>
                        {Math.round(current.temperature)}
                        {temp}
                    </div>
                    <div className={styles.currentLabel}>{info.label}</div>
                </div>
            </div>
            <div className={styles.currentMeta}>
                <div>
                    <span className={styles.metaKey}>Feels like</span> {Math.round(current.apparentTemperature)}
                    {temp}
                </div>
                <div>
                    <span className={styles.metaKey}>Humidity</span> {current.humidity}%
                </div>
                <div>
                    <span className={styles.metaKey}>Wind</span> {Math.round(current.windSpeed)}{" "}
                    {windSuffix(units.windSpeed)} {windCardinal(current.windDirection)}
                </div>
                <div>
                    <span className={styles.metaKey}>Precip</span> {current.precipitation}{" "}
                    {precipitationSuffix(units.precipitation)}
                </div>
            </div>
        </section>
    );
}
