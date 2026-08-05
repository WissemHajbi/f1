import type { PropsWithChildren } from 'react';
import { StyleSheet, Text, View } from 'react-native';

import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';

export function SectionTitle({ children, aside }: PropsWithChildren<{ aside?: string }>) {
  return <View style={styles.header}><Text style={styles.title}>{children}</Text>{aside ? <Text style={styles.aside}>{aside}</Text> : null}</View>;
}

const styles = StyleSheet.create({
  header: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: -spacing.md },
  title: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 21, letterSpacing: 0.7, textTransform: 'uppercase' },
  aside: { color: colors.textMuted, fontFamily: 'Inter_600SemiBold', fontSize: 10, letterSpacing: 1.2, textTransform: 'uppercase' },
});
