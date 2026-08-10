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
