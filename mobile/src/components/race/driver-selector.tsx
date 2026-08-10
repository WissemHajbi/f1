import { ScrollView, Pressable, StyleSheet, Text } from 'react-native';

import type { RaceDriverStats } from '@/api/types';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';

export function DriverSelector({ drivers, selected, onSelect }: {
  drivers: RaceDriverStats[];
  selected: number;
  onSelect: (driverNumber: number) => void;
}) {
  return (
    <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.list}>
      {drivers.map(({ driver }) => {
        const active = driver.driver_number === selected;
        const teamColor = /^[0-9a-f]{6}$/i.test(driver.team_colour) ? `#${driver.team_colour}` : colors.red;
        return <Pressable key={driver.driver_number} accessibilityRole="button" accessibilityState={{ selected: active }}
          accessibilityLabel={`Select ${driver.full_name}`} onPress={() => onSelect(driver.driver_number)}
          style={[styles.item, active && { borderColor: teamColor, backgroundColor: colors.surfaceRaised }]}>
          <Text style={[styles.number, active && { color: teamColor }]}>{driver.driver_number}</Text>
          <Text style={[styles.code, active && styles.activeCode]}>{driver.name_acronym}</Text>
        </Pressable>;
      })}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  list: { gap: spacing.sm, paddingRight: spacing.lg },
  item: { minWidth: 62, height: 56, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface, alignItems: 'center', justifyContent: 'center' },
  number: { color: colors.textDim, fontFamily: 'BarlowCondensed_700Bold', fontSize: 18, lineHeight: 19 },
  code: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 11, letterSpacing: 1 },
  activeCode: { color: colors.text },
});
