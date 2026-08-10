import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { router, useLocalSearchParams } from 'expo-router';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { DEFAULT_SEASON } from '@/api/client';
import { queries } from '@/api/hooks';
import { Card } from '@/components/common/card';
import { DataState } from '@/components/common/data-state';
import { PageHeader } from '@/components/layout/page-header';
import { Screen } from '@/components/layout/screen';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export default function DriverProfileScreen() {
  const number = Number(useLocalSearchParams<{ number: string }>().number);
  const query = useQuery(queries.drivers());
  const driver = query.data?.find((item) => item.driver_number === number);
  const teamColor = driver && /^[0-9a-f]{6}$/i.test(driver.team_colour) ? `#${driver.team_colour}` : colors.red;

  return (
    <Screen refreshing={query.isRefetching} onRefresh={() => query.refetch()}>
      <PageHeader eyebrow={`${DEFAULT_SEASON} DRIVER`} title={driver?.name_acronym || `#${number}`}
        action={<Pressable accessibilityRole="button" accessibilityLabel="Go back" onPress={() => router.back()} style={styles.back}>
          <MaterialCommunityIcons name="arrow-left" color={colors.text} size={22} />
        </Pressable>} />
      <DataState loading={query.isLoading} error={query.error} empty={!query.isLoading && !driver} onRetry={() => query.refetch()} />

      {driver ? <>
        <Card accent={teamColor} style={styles.hero}>
          <View style={styles.numberBlock}><Text style={[styles.number, { color: teamColor }]}>{driver.driver_number}</Text></View>
          <View style={styles.identity}>
            <Text style={styles.firstName}>{driver.first_name}</Text>
            <Text style={styles.lastName}>{driver.last_name}</Text>
            <Text style={[styles.team, { color: teamColor }]}>{driver.team_name}</Text>
          </View>
        </Card>
        <Card>
          <Detail label="Broadcast name" value={driver.broadcast_name} />
          <Detail label="Abbreviation" value={driver.name_acronym} divided />
          <Detail label="Country" value={driver.country_code || 'Unknown'} divided />
        </Card>
      </> : null}
    </Screen>
  );
}

function Detail({ label, value, divided = false }: { label: string; value: string; divided?: boolean }) {
  return <View style={[styles.detail, divided && styles.divider]}><Text style={typography.label}>{label}</Text><Text style={styles.value}>{value}</Text></View>;
}

const styles = StyleSheet.create({
  back: { width: 42, height: 42, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border },
  hero: { flexDirection: 'row', alignItems: 'center', gap: spacing.lg },
  numberBlock: { width: 94, minHeight: 112, borderRadius: radius.md, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.surfaceRaised },
  number: { fontFamily: 'BarlowCondensed_700Bold', fontSize: 58, lineHeight: 62 },
  identity: { flex: 1 },
  firstName: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 20, textTransform: 'uppercase' },
  lastName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 31, lineHeight: 33, textTransform: 'uppercase' },
  team: { fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 15, marginTop: spacing.sm, textTransform: 'uppercase' },
  detail: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  divider: { borderTopWidth: 1, borderTopColor: colors.border },
  value: { color: colors.text, fontFamily: 'Inter_600SemiBold', fontSize: 14, textAlign: 'right', flexShrink: 1 },
});
