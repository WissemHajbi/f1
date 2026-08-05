import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { router } from 'expo-router';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { useState } from 'react';

import { DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { Card } from '@/components/common/card';
import { DataState } from '@/components/common/data-state';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { typography } from '@/theme/typography';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';

const dateParts = (value?: string) => {
  if (!value) return { day: '--', month: 'TBC' };
  const date = new Date(value);
  return { day: String(date.getDate()).padStart(2, '0'), month: date.toLocaleDateString('en', { month: 'short' }).toUpperCase() };
};

export default function CalendarScreen() {
  const query = useQuery(queries.calendar());
  const [renderedAt] = useState(() => Date.now());
  return <Screen refreshing={query.isRefetching} onRefresh={() => query.refetch()}>
    <PageHeader eyebrow="WORLD CHAMPIONSHIP" title={`${DEFAULT_SEASON} Calendar`} />
    <DataState loading={query.isLoading} error={query.error} empty={query.data?.length === 0} onRetry={() => query.refetch()} />
    {query.data?.map((event) => {
      const date = dateParts(event.race_at);
      const complete = event.race_at ? new Date(event.race_at).getTime() < renderedAt : false;
      return <Pressable key={event.round} accessibilityRole="button" accessibilityLabel={`Open ${event.name} results`}
        onPress={() => router.push({ pathname: '/race/[round]', params: { round: event.round } })}>
        <Card accent={complete ? colors.borderStrong : colors.red} style={styles.raceCard}>
          <View style={styles.round}><Text style={styles.roundLabel}>R{String(event.round).padStart(2, '0')}</Text><Text style={styles.day}>{date.day}</Text><Text style={styles.month}>{date.month}</Text></View>
          <View style={styles.raceInfo}><Text style={styles.country}>{event.circuit.country}</Text><Text numberOfLines={1} style={styles.name}>{event.name}</Text><View style={styles.location}><MaterialCommunityIcons name="map-marker-outline" color={colors.textDim} size={14} /><Text style={typography.muted}>{event.circuit.locality} · {event.circuit.name}</Text></View></View>
          <MaterialCommunityIcons name="chevron-right" color={complete ? colors.textDim : colors.red} size={21} />
        </Card>
      </Pressable>;
    })}
  </Screen>;
}

const styles = StyleSheet.create({
  raceCard: { flexDirection: 'row', alignItems: 'center', gap: spacing.lg, paddingVertical: spacing.md },
  round: { width: 50, alignItems: 'center', borderRightWidth: 1, borderRightColor: colors.border, paddingRight: spacing.md },
  roundLabel: { color: colors.red, fontFamily: 'Inter_600SemiBold', fontSize: 9, letterSpacing: 1.2 },
  day: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 27, lineHeight: 29 },
  month: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 11, letterSpacing: 1 },
  raceInfo: { flex: 1 }, country: { color: colors.textMuted, fontFamily: 'Inter_600SemiBold', fontSize: 9, letterSpacing: 1.4, textTransform: 'uppercase' },
  name: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 21, textTransform: 'uppercase', marginTop: 1 },
  location: { flexDirection: 'row', alignItems: 'center', gap: 3, marginTop: 3 },
});
