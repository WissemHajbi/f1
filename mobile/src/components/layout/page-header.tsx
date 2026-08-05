import type { ReactNode } from 'react';
import { StyleSheet, Text, View } from 'react-native';

import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';

export function PageHeader({ eyebrow, title, action }: { eyebrow: string; title: string; action?: ReactNode }) {
  return <View style={styles.header}><View style={{ flex: 1 }}><Text style={styles.eyebrow}>{eyebrow}</Text><Text style={styles.title}>{title}</Text></View>{action}</View>;
}

const styles = StyleSheet.create({
  header: { minHeight: 92, flexDirection: 'row', alignItems: 'center', paddingTop: spacing.md },
  eyebrow: { color: colors.red, fontFamily: 'Inter_600SemiBold', fontSize: 10, letterSpacing: 2.6, marginBottom: 3 },
  title: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 38, letterSpacing: 0.3, lineHeight: 40, textTransform: 'uppercase' },
});
