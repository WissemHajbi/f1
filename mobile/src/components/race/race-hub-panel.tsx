import { StyleSheet, Text, View } from 'react-native';

import type { RaceHub } from '@/api/types';
import { Card } from '@/components/common/card';
import { SectionTitle } from '@/components/common/section-title';
import { DriverSelector } from '@/components/race/driver-selector';
import { TrackMap } from '@/components/race/track-map';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export function RaceHubPanel({ hub, onSelectDriver }: { hub: RaceHub; onSelectDriver: (number: number) => void }) {
  const stats = hub.drivers.find((item) => item.driver.driver_number === hub.selected_driver_number) ?? hub.drivers[0];
  const bestLaps = hub.laps.filter((lap) => lap.duration && !lap.is_pit_out_lap).sort((a, b) => a.duration! - b.duration!).slice(0, 5);
  if (!stats) return null;
  return <>
    <SectionTitle aside={`Session ${hub.session_key}`}>Circuit</SectionTitle>
    <TrackMap points={hub.track} sourceLap={hub.track_source_lap} />

    <SectionTitle aside={`${hub.drivers.length} drivers`}>Explore driver</SectionTitle>
    <DriverSelector drivers={hub.drivers} selected={hub.selected_driver_number} onSelect={onSelectDriver} />

    <Card accent={teamColor(stats.driver.team_colour)}>
      <Text style={typography.label}>{stats.driver.team_name}</Text>
      <Text style={styles.driverName}>{stats.driver.first_name} <Text style={styles.lastName}>{stats.driver.last_name}</Text></Text>
      <View style={styles.statsGrid}>
        <Stat label="Best lap" value={formatDuration(stats.best_lap_duration)} />
        <Stat label="Top speed" value={stats.top_speed ? `${stats.top_speed} km/h` : '—'} />
        <Stat label="Pit stops" value={String(stats.pit_stops)} />
        <Stat label="Overtakes" value={String(stats.overtakes)} />
        <Stat label="Tyre stints" value={String(stats.stints)} />
        <Stat label="Radio" value={String(stats.radio_messages)} />
      </View>
    </Card>

    <SectionTitle aside={`${stats.completed_laps} laps`}>Fastest laps</SectionTitle>
    <Card>
      {bestLaps.map((lap, index) => <View key={lap.lap_number} style={[styles.lap, index > 0 && styles.divider]}>
        <Text style={styles.lapNumber}>L{lap.lap_number}</Text>
        <View style={styles.lapBar}><View style={[styles.lapFill, { width: `${lapBarWidth(lap.duration!, bestLaps.map((item) => item.duration!))}%` }]} /></View>
        <Text style={styles.lapTime}>{formatDuration(lap.duration)}</Text>
      </View>)}
      {!bestLaps.length ? <Text style={typography.muted}>No synchronized lap timing.</Text> : null}
    </Card>
  </>;
}

function Stat({ label, value }: { label: string; value: string }) {
  return <View style={styles.stat}><Text style={typography.label}>{label}</Text><Text style={styles.statValue}>{value}</Text></View>;
}

export function formatDuration(seconds?: number) {
  if (seconds === undefined) return '—';
  const minutes = Math.floor(seconds / 60);
  return `${minutes}:${(seconds % 60).toFixed(3).padStart(6, '0')}`;
}

function teamColor(value: string) { return /^[0-9a-f]{6}$/i.test(value) ? `#${value}` : colors.red; }
function lapBarWidth(value: number, values: number[]) {
  const fastest = Math.min(...values), slowest = Math.max(...values);
  if (fastest === slowest) return 100;
  return 70 + ((slowest - value) / (slowest - fastest)) * 30;
}

const styles = StyleSheet.create({
  driverName: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 27, textTransform: 'uppercase', marginTop: 2 },
  lastName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold' },
  statsGrid: { flexDirection: 'row', flexWrap: 'wrap', marginTop: spacing.md, borderTopWidth: 1, borderTopColor: colors.border },
  stat: { width: '50%', paddingVertical: spacing.md, paddingRight: spacing.sm },
  statValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 20, marginTop: 2 },
  lap: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, paddingVertical: spacing.sm },
  divider: { borderTopWidth: 1, borderTopColor: colors.border },
  lapNumber: { width: 32, color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 13 },
  lapBar: { flex: 1, height: 5, borderRadius: radius.pill, backgroundColor: colors.surfaceRaised, overflow: 'hidden' },
  lapFill: { height: '100%', borderRadius: radius.pill, backgroundColor: colors.red },
  lapTime: { width: 68, color: colors.text, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 14, textAlign: 'right' },
});
