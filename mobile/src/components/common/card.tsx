import type { PropsWithChildren } from 'react';
import { StyleSheet, View, type ViewStyle } from 'react-native';

import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';

export function Card({ children, accent, style }: PropsWithChildren<{ accent?: string; style?: ViewStyle | ViewStyle[] }>) {
  return <View style={[styles.card, style]}>{accent ? <View style={[styles.accent, { backgroundColor: accent }]} /> : null}{children}</View>;
}

const styles = StyleSheet.create({
  card: { overflow: 'hidden', backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: spacing.lg },
  accent: { position: 'absolute', left: 0, top: 0, bottom: 0, width: 3 },
});
