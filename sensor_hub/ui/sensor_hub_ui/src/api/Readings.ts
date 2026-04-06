import {get} from "./Client.ts";
import type {Reading} from "../types/types.ts";

export const ReadingsApi = {
  getBetweenDates: (start: string, end: string, sensor?: string, measurementType?: string) => {
    const params = new URLSearchParams({start, end});
    if (sensor) params.set('sensor', sensor);
    if (measurementType) params.set('type', measurementType);
    return get<Reading[]>(`/readings/between?${params.toString()}`);
  },
  getBetweenDatesHourly: (start: string, end: string, sensor?: string, measurementType?: string) => {
    const params = new URLSearchParams({start, end});
    if (sensor) params.set('sensor', sensor);
    if (measurementType) params.set('type', measurementType);
    return get<Reading[]>(`/readings/hourly/between?${params.toString()}`);
  },
}
