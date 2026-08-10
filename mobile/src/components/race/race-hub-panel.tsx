import { useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';

import type { RaceHub } from '@/api/types';
import { Card } from '@/components/common/card';
import { SectionTitle } from '@/components/common/section-title';
import { DriverSelector } from '@/components/race/driver-selector';
import { RadioPlayer } from '@/components/race/radio-player';
import { TrackMap } from '@/components/race/track-map';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

type Panel = 'pace' | 'strategy' | 'radio';

export function RaceHubPanel({ hub, selectedDriverNumber, selectedLapNumber, updating, onSelectDriver, onSelectLap }: {
  hub: RaceHub;
  selectedDriverNumber: number;
  selectedLapNumber?: number;
  updating: boolean;
  onSelectDriver: (number: number) => void;
  onSelectLap: (number: number) => void;
}) {
  const [panel, setPanel] = useState<Panel>('pace');
  const [kpiInfo, setKpiInfo] = useState('Tap any statistic to learn what it means and open its detail panel.');
  const stats = hub.drivers.find((item) => item.driver.driver_number === selectedDriverNumber) ?? hub.drivers[0];
  const detailsReady = hub.selected_driver_number === selectedDriverNumber;
  const activeLap = selectedLapNumber ?? stats?.best_lap_number;
  const traceReady = detailsReady && hub.driver_trace_lap === activeLap;
  if (!stats) return null;
  const accent = teamColor(stats.driver.team_colour);

  return <>
    <SectionTitle aside={`Session ${hub.session_key}`}>Circuit</SectionTitle>
    <TrackMap points={hub.track} trace={traceReady ? hub.driver_trace : []} traceColor={accent}
      sourceLap={hub.track_source_lap} traceLap={traceReady ? hub.driver_trace_lap : activeLap}
      attribution={hub.track_attribution} accuracy={hub.track_accuracy} estimatedWidth={hub.track_estimated_width_m} />
    {updating && !traceReady ? <Text style={styles.updating}>Loading #{selectedDriverNumber} lap {activeLap} racing line…</Text> : null}

    <SectionTitle aside={`${hub.drivers.length} classified`}>Choose driver</SectionTitle>
    <DriverSelector drivers={hub.drivers} selected={selectedDriverNumber} onSelect={onSelectDriver} />

    <Card accent={accent}>
      <Text style={typography.label}>{stats.driver.team_name}</Text>
      <Text style={styles.driverName}>{stats.driver.first_name} <Text style={styles.lastName}>{stats.driver.last_name}</Text></Text>
      <View style={styles.statsGrid}>
        <Stat label="Best lap" value={formatDuration(stats.best_lap_duration)} active={panel === 'pace'} description={KPI_HELP.bestLap} onPress={() => { setPanel('pace'); setKpiInfo(KPI_HELP.bestLap); }} />
        <Stat label="Top speed" value={stats.top_speed ? `${stats.top_speed} km/h` : '—'} active={panel === 'pace'} description={KPI_HELP.topSpeed} onPress={() => { setPanel('pace'); setKpiInfo(KPI_HELP.topSpeed); }} />
        <Stat label="Pit stops" value={String(stats.pit_stops)} active={panel === 'strategy'} description={KPI_HELP.pitStops} onPress={() => { setPanel('strategy'); setKpiInfo(KPI_HELP.pitStops); }} />
        <Stat label="Overtakes" value={String(stats.overtakes)} active={panel === 'strategy'} description={KPI_HELP.overtakes} onPress={() => { setPanel('strategy'); setKpiInfo(KPI_HELP.overtakes); }} />
        <Stat label="Tyre stints" value={String(stats.stints)} active={panel === 'strategy'} description={KPI_HELP.stints} onPress={() => { setPanel('strategy'); setKpiInfo(KPI_HELP.stints); }} />
        <Stat label="Radio" value={String(stats.radio_messages)} active={panel === 'radio'} description={KPI_HELP.radio} onPress={() => { setPanel('radio'); setKpiInfo(KPI_HELP.radio); }} />
      </View>
      <View style={styles.kpiHelp}><Text style={styles.infoMark}>?</Text><Text style={styles.kpiHelpText}>{kpiInfo}</Text></View>
    </Card>

    <View style={styles.tabs}>
      {(['pace', 'strategy', 'radio'] as const).map((value) => <Pressable key={value} accessibilityRole="tab"
        accessibilityState={{ selected: panel === value }} onPress={() => setPanel(value)} style={[styles.tab, panel === value && { borderColor: accent }]}>
        <Text style={[styles.tabText, panel === value && { color: accent }]}>{value}</Text>
      </Pressable>)}
    </View>

    {panel === 'pace' ? <PacePanel key={selectedDriverNumber} stats={stats} laps={detailsReady ? hub.laps : []}
      selectedNumber={activeLap} ready={detailsReady} onSelectLap={onSelectLap} /> : null}
    {panel === 'strategy' ? <StrategyPanel stats={stats} /> : null}
    {panel === 'radio' ? <Card><RadioPlayer sessionKey={hub.session_key} driverNumber={selectedDriverNumber}
      count={stats.radio_messages} color={accent} /></Card> : null}
  </>;
}

function PacePanel({ stats, laps, selectedNumber, ready, onSelectLap }: {
  stats: RaceHub['drivers'][number];
  laps: RaceHub['laps'];
  selectedNumber?: number;
  ready: boolean;
  onSelectLap: (number: number) => void;
}) {
  const selected = laps.find((lap) => lap.lap_number === selectedNumber);
  const bestLaps = laps.filter((lap) => lap.duration && !lap.is_pit_out_lap).sort((a, b) => a.duration! - b.duration!).slice(0, 5);
  return <>
    <SectionTitle aside={`Lap ${stats.best_lap_number ?? '—'}`}>Best sectors</SectionTitle>
    <Text style={styles.explainer}>{"A circuit is split into three sectors. Lower times are faster; these are the driver's quickest values in each sector."}</Text>
    <View style={styles.sectors}>
      <Sector label="S1" value={stats.best_sector_1} />
      <Sector label="S2" value={stats.best_sector_2} />
      <Sector label="S3" value={stats.best_sector_3} />
    </View>

    <SectionTitle aside={`${laps.length} recorded`}>Inspect any lap</SectionTitle>
    {!ready ? <Card><Text style={typography.muted}>Updating lap detail…</Text></Card> : <>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.lapPicker}>
        {laps.map((lap) => <Pressable key={lap.lap_number} accessibilityRole="button"
          accessibilityLabel={`Show lap ${lap.lap_number}`} accessibilityState={{ selected: selectedNumber === lap.lap_number }}
          onPress={() => onSelectLap(lap.lap_number)} style={[styles.lapChip, selectedNumber === lap.lap_number && styles.activeLapChip]}>
          <Text style={[styles.lapChipText, selectedNumber === lap.lap_number && styles.activeLapChipText]}>L{lap.lap_number}</Text>
        </Pressable>)}
      </ScrollView>
      {selected ? <Card accent={selected.lap_number === stats.best_lap_number ? colors.green : colors.borderStrong}>
        <View style={styles.lapHeadline}><View><Text style={typography.label}>LAP {selected.lap_number}</Text><Text style={styles.selectedLapTime}>{formatDuration(selected.duration)}</Text></View>
          <Text style={styles.lapTag}>{selected.is_pit_out_lap ? 'PIT-OUT LAP' : selected.lap_number === stats.best_lap_number ? 'PERSONAL BEST' : 'RACE LAP'}</Text></View>
        <View style={styles.lapDetails}>
          <LapValue label="Sector 1" value={seconds(selected.sector_1_duration)} />
          <LapValue label="Sector 2" value={seconds(selected.sector_2_duration)} />
          <LapValue label="Sector 3" value={seconds(selected.sector_3_duration)} />
          <LapValue label="Speed trap" value={selected.speed_trap ? `${selected.speed_trap} km/h` : '—'} />
        </View>
      </Card> : null}
    </>}

    <SectionTitle>Five fastest laps</SectionTitle>
    <Card>
      {bestLaps.map((lap, index) => <Pressable key={lap.lap_number} accessibilityRole="button" onPress={() => onSelectLap(lap.lap_number)}
        style={[styles.lap, index > 0 && styles.divider]}>
        <Text style={styles.lapNumber}>L{lap.lap_number}</Text>
        <View style={styles.lapBar}><View style={[styles.lapFill, { width: `${lapBarWidth(lap.duration!, bestLaps.map((item) => item.duration!))}%` }]} /></View>
        <Text style={styles.lapTime}>{formatDuration(lap.duration)}</Text>
      </Pressable>)}
      {ready && !bestLaps.length ? <Text style={typography.muted}>No synchronized lap timing.</Text> : null}
    </Card>
  </>;
}

function StrategyPanel({ stats }: { stats: RaceHub['drivers'][number] }) {
  return <>
    <SectionTitle>Race activity</SectionTitle>
    <Text style={styles.explainer}>Strategy describes how the driver used tyres, visited the pits, and gained places during the race.</Text>
    <View style={styles.activity}>
      <Activity value={stats.completed_laps} label="Completed laps" />
      <Activity value={stats.stints} label="Tyre stints" />
      <Activity value={stats.pit_stops} label="Pit stops" />
      <Activity value={stats.overtakes} label="Overtakes" />
    </View>
  </>;
}

function Stat({ label, value, active, description, onPress }: { label: string; value: string; active: boolean; description: string; onPress: () => void }) {
  return <Pressable accessibilityRole="button" accessibilityLabel={`${label}: ${value}`} accessibilityHint={description}
    onPress={onPress} style={[styles.stat, active && styles.activeStat]}>
    <Text style={typography.label}>{label}</Text><Text style={styles.statValue}>{value}</Text>
  </Pressable>;
}
function Sector({ label, value }: { label: string; value?: number }) {
  return <Card style={styles.sector}><Text style={typography.label}>{label}</Text><Text style={styles.sectorValue}>{value ? `${value.toFixed(3)}s` : '—'}</Text></Card>;
}
function Activity({ value, label }: { value: number; label: string }) {
  return <Card style={styles.activityCard}><Text style={styles.activityValue}>{value}</Text><Text style={typography.label}>{label}</Text></Card>;
}
function LapValue({ label, value }: { label: string; value: string }) {
  return <View style={styles.lapValue}><Text style={typography.label}>{label}</Text><Text style={styles.lapValueText}>{value}</Text></View>;
}
const KPI_HELP = {
  bestLap: 'The shortest time this driver needed to complete one full lap.',
  topSpeed: 'The highest speed recorded at an official speed-measurement point on the circuit.',
  pitStops: 'Times the driver entered the pit lane for tyres, repairs, or penalties.',
  overtakes: 'Recorded moments where this driver passed another car for position.',
  stints: 'Separate runs on one set of tyres. A pit stop normally starts a new stint.',
  radio: 'Audio messages exchanged between the driver and the team.',
};
function seconds(value?: number) { return value === undefined ? '—' : `${value.toFixed(3)}s`; }
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
  updating: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 10, textAlign: 'center', marginTop: -spacing.md },
  driverName: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 27, textTransform: 'uppercase', marginTop: 2 },
  lastName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold' },
  statsGrid: { flexDirection: 'row', flexWrap: 'wrap', marginTop: spacing.md, borderTopWidth: 1, borderTopColor: colors.border },
  stat: { width: '50%', paddingVertical: spacing.md, paddingRight: spacing.sm, borderRadius: radius.sm },
  activeStat: { backgroundColor: colors.surfaceRaised },
  statValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 20, marginTop: 2 },
  kpiHelp: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, marginTop: spacing.sm, paddingTop: spacing.md, borderTopWidth: 1, borderTopColor: colors.border },
  infoMark: { width: 22, height: 22, borderRadius: radius.pill, backgroundColor: colors.surfaceRaised, color: colors.red, textAlign: 'center', lineHeight: 22, fontFamily: 'Inter_600SemiBold', fontSize: 12 },
  kpiHelpText: { flex: 1, color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 11, lineHeight: 16 },
  tabs: { flexDirection: 'row', gap: spacing.sm },
  tab: { flex: 1, borderRadius: radius.pill, borderWidth: 1, borderColor: colors.border, paddingVertical: spacing.sm, alignItems: 'center', backgroundColor: colors.surface },
  tabText: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 12, letterSpacing: 1, textTransform: 'uppercase' },
  explainer: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 11, lineHeight: 17, marginTop: -spacing.md },
  sectors: { flexDirection: 'row', gap: spacing.sm },
  sector: { flex: 1 },
  sectorValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 18, marginTop: spacing.xs },
  lapPicker: { gap: spacing.xs, paddingRight: spacing.lg },
  lapChip: { minWidth: 45, height: 38, borderRadius: radius.pill, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface, alignItems: 'center', justifyContent: 'center' },
  activeLapChip: { borderColor: colors.red, backgroundColor: colors.surfaceRaised },
  lapChipText: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 12 },
  activeLapChipText: { color: colors.text },
  lapHeadline: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  selectedLapTime: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 29, marginTop: 2 },
  lapTag: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 10, letterSpacing: 1 },
  lapDetails: { flexDirection: 'row', flexWrap: 'wrap', marginTop: spacing.md, borderTopWidth: 1, borderTopColor: colors.border },
  lapValue: { width: '50%', paddingTop: spacing.md },
  lapValueText: { color: colors.text, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 17, marginTop: 2 },
  lap: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, paddingVertical: spacing.sm },
  divider: { borderTopWidth: 1, borderTopColor: colors.border },
  lapNumber: { width: 32, color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 13 },
  lapBar: { flex: 1, height: 5, borderRadius: radius.pill, backgroundColor: colors.surfaceRaised, overflow: 'hidden' },
  lapFill: { height: '100%', borderRadius: radius.pill, backgroundColor: colors.red },
  lapTime: { width: 68, color: colors.text, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 14, textAlign: 'right' },
  activity: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  activityCard: { width: '48%', minHeight: 94, justifyContent: 'center' },
  activityValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 34 },
});
