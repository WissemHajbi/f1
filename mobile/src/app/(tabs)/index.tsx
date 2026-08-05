import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { LinearGradient } from 'expo-linear-gradient';
import { StyleSheet, Text, View } from 'react-native';

import { api, DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { Card } from '@/components/common/card';
import { DataState } from '@/components/common/data-state';
import { PositionBadge } from '@/components/common/position-badge';
import { SectionTitle } from '@/components/common/section-title';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { typography } from '@/theme/typography';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';

export default function HomeScreen() {
  const calendar = useQuery(queries.calendar());
  const standings = useQuery(queries.driverStandings());
  const latest = useQuery({ queryKey: ['results', 'latest'], queryFn: api.latestResult, staleTime: 60 * 60_000 });
  const loading = calendar.isLoading || standings.isLoading || latest.isLoading;
  const error = calendar.error || standings.error || latest.error;
  const race = latest.data;
  const winner = race?.results[0];

  return (
    <Screen refreshing={calendar.isRefetching || standings.isRefetching || latest.isRefetching}
      onRefresh={() => { calendar.refetch(); standings.refetch(); latest.refetch(); }}>
      <PageHeader eyebrow={`OIDYSTS / ${DEFAULT_SEASON}`} title="Race Control" action={<View style={styles.live}><View style={styles.liveDot} /><Text style={styles.liveText}>LOCAL DB</Text></View>} />
      {loading || error ? <DataState loading={loading} error={error as Error} onRetry={() => { calendar.refetch(); standings.refetch(); latest.refetch(); }} /> : <>
        {race && winner ? <LinearGradient colors={['#2A0810', '#111318']} start={{ x: 0, y: 0 }} end={{ x: 1, y: 1 }} style={styles.hero}>
            <View style={styles.heroTop}><Text style={styles.kicker}>LATEST CLASSIFICATION</Text><MaterialCommunityIcons name="flag-checkered" color={colors.textMuted} size={20} /></View>
            <Text style={styles.raceName}>{race.name}</Text>
            <Text style={styles.circuit}>{race.circuit.locality.toUpperCase()} · ROUND {String(race.round).padStart(2, '0')}</Text>
            <View style={styles.winnerRow}><PositionBadge position={1} /><View style={{ flex: 1 }}><Text style={styles.winnerName}>{winner.given_name} <Text style={{ color: colors.white }}>{winner.family_name}</Text></Text><Text style={typography.muted}>{winner.constructor.name} · {winner.time || winner.status}</Text></View></View>
            <View style={styles.heroLine} />
          </LinearGradient> : null}

        <SectionTitle aside={`After round ${latest.data?.round || '—'}`}>Championship</SectionTitle>
        <Card style={{ paddingVertical: spacing.sm }}>
          {standings.data?.slice(0, 3).map((driver, index) => <View key={driver.driver_id} style={[styles.standingRow, index < 2 && styles.rowBorder]}>
            <PositionBadge position={driver.position} />
            <View style={styles.nameBlock}><Text style={styles.driverCode}>{driver.code || driver.family_name.slice(0, 3)}</Text><Text style={typography.muted}>{driver.constructors[0]?.name}</Text></View>
            <View style={styles.points}><Text style={styles.pointsValue}>{driver.points}</Text><Text style={typography.label}>PTS</Text></View>
          </View>)}
        </Card>

        <SectionTitle aside={`${calendar.data?.length || 0} rounds`}>Season pulse</SectionTitle>
        <View style={styles.metrics}>
          <Card style={styles.metric}><MaterialCommunityIcons name="trophy-outline" color={colors.red} size={22} /><Text style={styles.metricValue}>{standings.data?.[0]?.wins || 0}</Text><Text style={typography.label}>LEADER WINS</Text></Card>
          <Card style={styles.metric}><MaterialCommunityIcons name="speedometer" color={colors.red} size={22} /><Text style={styles.metricValue}>{race?.results.length || 0}</Text><Text style={typography.label}>CLASSIFIED</Text></Card>
        </View>
      </>}
    </Screen>
  );
}

const styles = StyleSheet.create({
  live: { flexDirection: 'row', alignItems: 'center', gap: 7, borderWidth: 1, borderColor: colors.border, borderRadius: radius.pill, paddingHorizontal: 10, paddingVertical: 7 },
  liveDot: { width: 6, height: 6, borderRadius: 3, backgroundColor: colors.green },
  liveText: { color: colors.textMuted, fontFamily: 'Inter_600SemiBold', fontSize: 9, letterSpacing: 1 },
  hero: { borderRadius: radius.lg, borderWidth: 1, borderColor: '#4A1823', padding: spacing.xl, minHeight: 260, overflow: 'hidden' },
  heroTop: { flexDirection: 'row', justifyContent: 'space-between' },
  kicker: { color: colors.red, fontFamily: 'Inter_600SemiBold', fontSize: 10, letterSpacing: 2 },
  raceName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 39, lineHeight: 40, textTransform: 'uppercase', marginTop: spacing.xl, maxWidth: '85%' },
  circuit: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 13, letterSpacing: 1.5, marginTop: 5 },
  winnerRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, marginTop: spacing.xl },
  winnerName: { color: colors.textMuted, fontFamily: 'BarlowCondensed_700Bold', fontSize: 20, textTransform: 'uppercase' },
  heroLine: { position: 'absolute', height: 3, width: '54%', backgroundColor: colors.red, bottom: 0, left: 0 },
  standingRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  rowBorder: { borderBottomWidth: 1, borderBottomColor: colors.border },
  nameBlock: { flex: 1 }, driverCode: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 21 },
  points: { alignItems: 'flex-end' }, pointsValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 24 },
  metrics: { flexDirection: 'row', gap: spacing.md }, metric: { flex: 1, gap: spacing.sm }, metricValue: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 34 },
});
