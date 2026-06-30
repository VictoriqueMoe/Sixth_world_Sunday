import type { EventFrequency } from "../types/api";
import { parseServerDate } from "./time";

const WEEKDAYS = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
const ORDINALS = ["first", "second", "third", "fourth", "fifth"];

export interface FrequencyOption {
    value: EventFrequency;
    label: string;
}

function nthWeekdayIndex(date: Date): number {
    return Math.floor((date.getDate() - 1) / 7);
}

export function frequencyOptions(start: Date | null): FrequencyOption[] {
    if (!start || Number.isNaN(start.getTime())) {
        return [
            { value: "none", label: "Does not repeat" },
            { value: "weekly", label: "Weekly" },
            { value: "biweekly", label: "Every other week" },
            { value: "monthly", label: "Monthly" },
            { value: "annually", label: "Annually" },
        ];
    }

    const weekday = WEEKDAYS[start.getDay()];
    const ordinal = ORDINALS[nthWeekdayIndex(start)] ?? "last";
    const monthDay = start.toLocaleDateString(undefined, { month: "long", day: "numeric" });

    return [
        { value: "none", label: "Does not repeat" },
        { value: "weekly", label: `Weekly on ${weekday}` },
        { value: "biweekly", label: `Every other ${weekday}` },
        { value: "monthly", label: `Monthly on the ${ordinal} ${weekday}` },
        { value: "annually", label: `Annually on ${monthDay}` },
    ];
}

export function frequencyLabel(freq: EventFrequency, startStr: string): string {
    if (freq === "none") {
        return "";
    }

    const start = parseServerDate(startStr);
    const option = frequencyOptions(start).find(o => o.value === freq);
    return option ? `Repeats ${option.label.toLowerCase()}` : "";
}
