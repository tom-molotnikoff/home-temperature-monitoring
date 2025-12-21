import {get, patch} from "./Client.ts";
import type {PropertiesApiStructure} from "../types/types.ts";

export const PropertiesApi = {
    getProperties: () => get<PropertiesApiStructure>("/properties/"),
    patchProperties: (properties: PropertiesApiStructure) => patch<PropertiesApiStructure>("/properties/", properties)
}