import { keepPreviousData } from '@tanstack/react-query';

import { api, DEFAULT_SEASON } from './client';

export const queries = {
  calendar: (season = DEFAULT_SEASON) => ({
    queryKey: ['calendar', season],
    queryFn: () => api.calendar(season),
    staleTime: 60 * 60_000,
  }),
  drivers: (season = DEFAULT_SEASON) => ({
    queryKey: ['drivers', season],
    queryFn: () => api.drivers(season),
    staleTime: 60 * 60_000,
  }),
  driverStandings: (season = DEFAULT_SEASON) => ({
    queryKey: ['standings', 'drivers', season],
    queryFn: () => api.driverStandings(season),
    staleTime: 10 * 60_000,
  }),
  constructorStandings: (season = DEFAULT_SEASON) => ({
    queryKey: ['standings', 'constructors', season],
    queryFn: () => api.constructorStandings(season),
    staleTime: 10 * 60_000,
  }),
  raceResult: (round: number, season = DEFAULT_SEASON) => ({
    queryKey: ['results', season, round],
    queryFn: () => api.raceResult(round, season),
    staleTime: 60 * 60_000,
    enabled: Number.isInteger(round) && round > 0,
  }),
  raceHub: (round: number, driverNumber?: number, lapNumber?: number, season = DEFAULT_SEASON) => ({
    queryKey: ['race-hub', season, round, driverNumber ?? 'default', lapNumber ?? 'best'],
    queryFn: () => api.raceHub(round, driverNumber, lapNumber, season),
    staleTime: 10 * 60_000,
    enabled: Number.isInteger(round) && round > 0,
    placeholderData: keepPreviousData,
  }),
  teamRadio: (sessionKey: number, driverNumber: number, enabled = true) => ({
    queryKey: ['team-radio', sessionKey, driverNumber],
    queryFn: () => api.teamRadio(sessionKey, driverNumber),
    staleTime: 60 * 60_000,
    enabled: enabled && sessionKey > 0 && driverNumber > 0,
  }),
};
