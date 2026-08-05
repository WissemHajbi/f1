import { Platform } from 'react-native';

import type { ApiEnvelope, CalendarEvent, ConstructorStanding, DriverStanding, RaceClassification } from './types';

const localURL = Platform.select({ android: 'http://10.0.2.2:8080', default: 'http://localhost:8080' });
export const API_URL = (process.env.EXPO_PUBLIC_API_URL || localURL).replace(/\/$/, '');
export const DEFAULT_SEASON = Number(process.env.EXPO_PUBLIC_SEASON || 2025);

class ApiError extends Error {
  constructor(message: string, readonly status: number) {
    super(message);
  }
}

async function get<T>(path: string): Promise<T> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 12_000);
  try {
    const response = await fetch(`${API_URL}${path}`, { headers: { Accept: 'application/json' }, signal: controller.signal });
    const body = await response.json().catch(() => null) as { error?: string } | null;
    if (!response.ok) throw new ApiError(body?.error || `Request failed (${response.status})`, response.status);
    return body as T;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    if (error instanceof Error && error.name === 'AbortError') throw new Error('The API took too long to respond.');
    throw new Error(`Cannot reach the F1 API at ${API_URL}.`);
  } finally {
    clearTimeout(timer);
  }
}

const seasonQuery = (season: number) => `?season=${encodeURIComponent(season)}`;

export const api = {
  calendar: (season = DEFAULT_SEASON) =>
    get<ApiEnvelope<CalendarEvent[]>>(`/v1/calendar${seasonQuery(season)}`).then((value) => value.data),
  driverStandings: (season = DEFAULT_SEASON) =>
    get<ApiEnvelope<DriverStanding[]>>(`/v1/standings/drivers${seasonQuery(season)}`).then((value) => value.data),
  constructorStandings: (season = DEFAULT_SEASON) =>
    get<ApiEnvelope<ConstructorStanding[]>>(`/v1/standings/constructors${seasonQuery(season)}`).then((value) => value.data),
  latestResult: () => get<ApiEnvelope<RaceClassification>>('/v1/results/latest').then((value) => value.data),
  raceResult: (round: number, season = DEFAULT_SEASON) =>
    get<ApiEnvelope<RaceClassification[]>>(`/v1/results?season=${encodeURIComponent(season)}&round=${encodeURIComponent(round)}`)
      .then((value) => value.data[0]),
};
