import { StyleSheet, Text, View } from 'react-native';

import type { RaceResult } from '@/api/types';
import { PositionBadge } from '@/components/common/position-badge';
import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export function ClassificationRow({ result, divided }: { result: RaceResult; divided: boolean }) {
  return (
    <View style={[styles.row, divided && styles.divider]}>
      <PositionBadge position={result.position} />
      <View style={styles.identity}>
        <Text style={styles.name}>{result.given_name} <Text style={styles.familyName}>{result.family_name}</Text></Text>
        <Text style={typography.muted}>{result.constructor.name}</Text>
      </View>
      <View style={styles.classification}>
        <Text style={styles.status} numberOfLines={1}>{result.time || result.status}</Text>
        <Text style={typography.label}>{result.points} PTS</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingVertical: spacing.md },
  divider: { borderBottomWidth: 1, borderBottomColor: colors.border },
  identity: { flex: 1 },
  name: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 17, textTransform: 'uppercase' },
  familyName: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold' },
  classification: { maxWidth: 112, alignItems: 'flex-end' },
  status: { color: colors.text, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 14, maxWidth: 112 },
});
