import {get} from "./Client.ts";
import type {TemperatureReading} from "../types/types.ts";

export const TemperatureApi = {
  getBetweenDates: (start: string, end: string) => get<TemperatureReading[]>(`/temperature/readings/between?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`),
  getBetweenDatesHourly: (start: string, end: string) => get<TemperatureReading[]>(`/temperature/readings/hourly/between?start=${encodeURIComponent(start)}&end=${encodeURIComponent(end)}`),
}
