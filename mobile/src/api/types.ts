export type ApiEnvelope<T, M = Record<string, unknown>> = { data: T; meta?: M };

export type Circuit = {
  id: string;
  name: string;
  locality: string;
  country: string;
  latitude: number;
  longitude: number;
};

export type CalendarEvent = {
  season: number;
  round: number;
  name: string;
  race_at?: string;
  circuit: Circuit;
  sessions: { type: string; start_at?: string }[];
};

export type ConstructorRef = { id: string; name: string; nationality: string };

export type DriverStanding = {
  position: number;
  points: number;
  wins: number;
  driver_id: string;
  permanent_number?: string;
  code?: string;
  given_name: string;
  family_name: string;
  nationality: string;
  constructors: ConstructorRef[];
};

export type ConstructorStanding = {
  position: number;
  points: number;
  wins: number;
  constructor: ConstructorRef;
};

export type RaceResult = {
  position: number;
  points: number;
  status: string;
  time?: string;
  driver_code?: string;
  given_name: string;
  family_name: string;
  constructor: ConstructorRef;
};

export type RaceClassification = {
  season: number;
  round: number;
  name: string;
  race_at?: string;
  circuit: Circuit;
  results: RaceResult[];
};

export type Driver = {
  session_key: number;
  driver_number: number;
  broadcast_name: string;
  full_name: string;
  name_acronym: string;
  team_name: string;
  team_colour: string;
  first_name: string;
  last_name: string;
  country_code?: string | null;
};

export type RaceDriverStats = {
  driver: Driver;
  completed_laps: number;
  best_lap_number?: number;
  best_lap_duration?: number;
  best_sector_1?: number;
  best_sector_2?: number;
  best_sector_3?: number;
  top_speed?: number;
  pit_stops: number;
  stints: number;
  overtakes: number;
  radio_messages: number;
};

export type RaceLapStat = {
  lap_number: number;
  duration?: number;
  sector_1_duration?: number;
  sector_2_duration?: number;
  sector_3_duration?: number;
  speed_trap?: number;
  is_pit_out_lap: boolean;
};

export type RaceHub = {
  season: number;
  round: number;
  session_key: number;
  drivers: RaceDriverStats[];
  selected_driver_number: number;
  laps: RaceLapStat[];
  track: { x: number; y: number }[];
  track_source_driver?: number;
  track_source_lap?: number;
};
