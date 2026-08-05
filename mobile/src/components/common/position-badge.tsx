import { StyleSheet, Text, View } from 'react-native';

import { colors } from '@/theme/colors';
import { radius } from '@/theme/spacing';

export function PositionBadge({ position }: { position: number }) {
  return <View style={[styles.badge, position === 1 && styles.first]}><Text style={styles.text}>{String(position).padStart(2, '0')}</Text></View>;
}

const styles = StyleSheet.create({
  badge: { width: 40, height: 40, borderRadius: radius.sm, backgroundColor: colors.surfaceSoft, alignItems: 'center', justifyContent: 'center' },
  first: { backgroundColor: colors.red },
  text: { color: colors.white, fontFamily: 'BarlowCondensed_700Bold', fontSize: 20 },
});
