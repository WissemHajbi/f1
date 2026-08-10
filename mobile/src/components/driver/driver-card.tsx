import { MaterialCommunityIcons } from '@expo/vector-icons';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import type { Driver } from '@/api/types';
import { Card } from '@/components/common/card';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export function DriverCard({ driver, onPress }: { driver: Driver; onPress: () => void }) {
  const teamColor = /^([0-9a-f]{6})$/i.test(driver.team_colour) ? `#${driver.team_colour}` : colors.red;
  return (
    <Pressable accessibilityRole="button" accessibilityLabel={`Open ${driver.full_name}`} onPress={onPress}>
      <Card style={styles.card}>
        <View style={[styles.number, { borderColor: teamColor }]}>
          <Text style={styles.numberText}>{driver.driver_number}</Text>
        </View>
        <View style={styles.identity}>
          <Text style={styles.name}>{driver.first_name} <Text style={styles.lastName}>{driver.last_name}</Text></Text>
          <Text style={typography.muted}>{driver.team_name}</Text>
        </View>
        <Text style={[styles.code, { color: teamColor }]}>{driver.name_acronym}</Text>
        <MaterialCommunityIcons name="chevron-right" size={20} color={colors.textDim} />
      </Card>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  number: { width: 43, height: 43, borderRadius: radius.sm, borderLeftWidth: 3, backgroundColor: colors.surfaceRaised, alignItems: 'center', justifyContent: 'center' },
  numberText: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 22 },
  identity: { flex: 1 },
  name: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 17, textTransform: 'uppercase' },
  lastName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold' },
  code: { fontFamily: 'BarlowCondensed_700Bold', fontSize: 14, letterSpacing: 1 },
});
