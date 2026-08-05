import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { router, useLocalSearchParams } from 'expo-router';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { Card } from '@/components/common/card';
import { DataState } from '@/components/common/data-state';
import { SectionTitle } from '@/components/common/section-title';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { ClassificationRow } from '@/components/race/classification-row';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export default function RaceDetailsScreen() {
  const round = Number(useLocalSearchParams<{ round: string }>().round);
  const query = useQuery(queries.raceResult(round));
  const race = query.data;

  return (
    <Screen refreshing={query.isRefetching} onRefresh={() => query.refetch()}>
      <PageHeader eyebrow={`ROUND ${String(round).padStart(2, '0')} / ${DEFAULT_SEASON}`} title={race?.name || 'Race results'}
        action={<Pressable accessibilityRole="button" accessibilityLabel="Go back" onPress={() => router.back()} style={styles.back}>
          <MaterialCommunityIcons name="arrow-left" color={colors.text} size={22} />
        </Pressable>} />

      <DataState loading={query.isLoading} error={query.error} empty={!query.isLoading && !race} onRetry={() => query.refetch()} />

      {race ? <>
        <Card accent={colors.red} style={styles.summary}>
          <View style={styles.summaryCopy}>
            <Text style={typography.label}>{race.circuit.country}</Text>
            <Text style={styles.circuit}>{race.circuit.name}</Text>
            <View style={styles.location}><MaterialCommunityIcons name="map-marker-outline" color={colors.textMuted} size={14} /><Text style={typography.muted}>{race.circuit.locality}</Text></View>
          </View>
          <View style={styles.dateBlock}>
            <Text style={styles.date}>{formatRaceDate(race.race_at)}</Text>
            <Text style={typography.label}>RACE DATE</Text>
          </View>
        </Card>

        <SectionTitle aside={`${race.results.length} classified`}>Final classification</SectionTitle>
        <Card style={styles.results}>
          {race.results.map((result, index) => <ClassificationRow key={`${result.position}-${result.family_name}`}
            result={result} divided={index < race.results.length - 1} />)}
        </Card>
      </> : null}
    </Screen>
  );
}

function formatRaceDate(value?: string) {
  if (!value) return 'TBC';
  return new Date(value).toLocaleDateString('en', { day: '2-digit', month: 'short' }).toUpperCase();
}

const styles = StyleSheet.create({
  back: { width: 42, height: 42, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border },
  summary: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: spacing.md },
  summaryCopy: { flex: 1 },
  circuit: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 25, lineHeight: 27, textTransform: 'uppercase', marginTop: 2 },
  location: { flexDirection: 'row', alignItems: 'center', gap: 3, marginTop: spacing.xs },
  dateBlock: { alignItems: 'flex-end' },
  date: { color: colors.red, fontFamily: 'BarlowCondensed_700Bold', fontSize: 23 },
  results: { paddingVertical: spacing.xs },
});
