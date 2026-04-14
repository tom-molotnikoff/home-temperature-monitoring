import {get} from "./Client.ts";
import type {AggregatedReadingsResponse} from "../types/types.ts";

export const ReadingsApi = {
  getBetweenDates: (start: string, end: string, sensor?: string, measurementType?: string, aggregation?: string, aggregationFunction?: string) => {
    const params = new URLSearchParams({start, end});
    if (sensor) params.set('sensor', sensor);
    if (measurementType) params.set('type', measurementType);
    if (aggregation) params.set('aggregation', aggregation);
    if (aggregationFunction) params.set('aggregation_function', aggregationFunction);
    return get<AggregatedReadingsResponse>(`/readings/between?${params.toString()}`);
  },
}
