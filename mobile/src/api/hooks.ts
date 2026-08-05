import { api, DEFAULT_SEASON } from './client';

export const queries = {
  calendar: (season = DEFAULT_SEASON) => ({
    queryKey: ['calendar', season],
    queryFn: () => api.calendar(season),
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
};
