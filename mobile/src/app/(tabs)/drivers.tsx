import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { router } from 'expo-router';
import { useMemo, useState } from 'react';
import { StyleSheet, Text, TextInput, View } from 'react-native';

import { DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { DataState } from '@/components/common/data-state';
import { DriverCard } from '@/components/driver/driver-card';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export default function DriversScreen() {
  const [search, setSearch] = useState('');
  const query = useQuery(queries.drivers());
  const drivers = useMemo(() => {
    const term = search.trim().toLowerCase();
    if (!term) return query.data ?? [];
    return (query.data ?? []).filter((driver) =>
      `${driver.full_name} ${driver.name_acronym} ${driver.team_name} ${driver.driver_number}`.toLowerCase().includes(term));
  }, [query.data, search]);

  return (
    <Screen refreshing={query.isRefetching} onRefresh={() => query.refetch()}>
      <PageHeader eyebrow={`${DEFAULT_SEASON} GRID`} title="Drivers" />
      <View style={styles.search}>
        <MaterialCommunityIcons name="magnify" color={colors.textDim} size={20} />
        <TextInput value={search} onChangeText={setSearch} placeholder="Driver, number or team" placeholderTextColor={colors.textDim}
          autoCorrect={false} returnKeyType="search" style={styles.input} accessibilityLabel="Search drivers" />
      </View>

      <DataState loading={query.isLoading} error={query.error} empty={!query.isLoading && !query.data?.length} onRetry={() => query.refetch()} />
      {!query.isLoading && query.data?.length && !drivers.length ? <View style={styles.noMatches}>
        <Text style={typography.body}>No drivers match “{search.trim()}”.</Text>
      </View> : null}
      <View style={styles.list}>
        {drivers.map((driver) => <DriverCard key={driver.driver_number} driver={driver}
          onPress={() => router.push({ pathname: '/driver/[number]', params: { number: driver.driver_number } })} />)}
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  search: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, minHeight: 48, paddingHorizontal: spacing.md, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface },
  input: { flex: 1, color: colors.text, fontFamily: 'Inter_400Regular', fontSize: 14, paddingVertical: spacing.sm },
  list: { gap: spacing.sm },
  noMatches: { paddingVertical: spacing.xl, alignItems: 'center' },
});
