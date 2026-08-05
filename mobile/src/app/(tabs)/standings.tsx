import { useQuery } from '@tanstack/react-query';
import * as Haptics from 'expo-haptics';
import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { Card } from '@/components/common/card';
import { DataState } from '@/components/common/data-state';
import { PositionBadge } from '@/components/common/position-badge';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { typography } from '@/theme/typography';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';

export default function StandingsScreen() {
  const [mode, setMode] = useState<'drivers' | 'teams'>('drivers');
  const drivers = useQuery(queries.driverStandings());
  const teams = useQuery(queries.constructorStandings());
  const active = mode === 'drivers' ? drivers : teams;
  const changeMode = (next: typeof mode) => { Haptics.selectionAsync(); setMode(next); };

  return <Screen refreshing={active.isRefetching} onRefresh={() => active.refetch()}>
    <PageHeader eyebrow={`${DEFAULT_SEASON} CLASSIFICATION`} title="Standings" />
    <View style={styles.segment}>
      {(['drivers', 'teams'] as const).map((item) => <Pressable key={item} onPress={() => changeMode(item)} style={[styles.segmentItem, mode === item && styles.segmentActive]}><Text style={[styles.segmentText, mode === item && styles.segmentTextActive]}>{item}</Text></Pressable>)}
    </View>
    <DataState loading={active.isLoading} error={active.error} empty={active.data?.length === 0} onRetry={() => active.refetch()} />
    {mode === 'drivers' ? drivers.data?.map((driver) => <Card key={driver.driver_id} accent={driver.position <= 3 ? colors.red : undefined} style={styles.row}>
      <PositionBadge position={driver.position} /><View style={styles.person}><Text style={styles.first}>{driver.given_name}</Text><Text style={styles.last}>{driver.family_name}</Text><Text style={typography.muted}>{driver.constructors[0]?.name || 'Independent'}</Text></View>
      <View style={styles.score}><Text style={styles.points}>{driver.points}</Text><Text style={typography.label}>POINTS</Text><Text style={styles.wins}>{driver.wins} WINS</Text></View>
    </Card>) : teams.data?.map((team) => <Card key={team.constructor.id} accent={team.position <= 3 ? colors.red : undefined} style={styles.row}>
      <PositionBadge position={team.position} /><View style={styles.person}><Text style={styles.last}>{team.constructor.name}</Text><Text style={typography.muted}>{team.constructor.nationality}</Text></View>
      <View style={styles.score}><Text style={styles.points}>{team.points}</Text><Text style={typography.label}>POINTS</Text><Text style={styles.wins}>{team.wins} WINS</Text></View>
    </Card>)}
  </Screen>;
}

const styles = StyleSheet.create({
  segment: { flexDirection: 'row', backgroundColor: colors.surface, borderRadius: radius.pill, borderWidth: 1, borderColor: colors.border, padding: 4 },
  segmentItem: { flex: 1, alignItems: 'center', paddingVertical: 10, borderRadius: radius.pill },
  segmentActive: { backgroundColor: colors.red },
  segmentText: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 14, textTransform: 'uppercase', letterSpacing: 1 },
  segmentTextActive: { color: colors.white },
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  person: { flex: 1 }, first: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 14, textTransform: 'uppercase' },
  last: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 22, textTransform: 'uppercase', lineHeight: 23 },
  score: { alignItems: 'flex-end' }, points: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 27, lineHeight: 29 },
  wins: { color: colors.red, fontFamily: 'Inter_600SemiBold', fontSize: 8, letterSpacing: 1, marginTop: 3 },
});
