import { MaterialCommunityIcons } from '@expo/vector-icons';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';

import { API_URL } from '@/api/client';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';

export function DataState({ loading, error, empty, onRetry }: { loading?: boolean; error?: Error | null; empty?: boolean; onRetry?: () => void }) {
  if (loading) return <View style={styles.state}><ActivityIndicator color={colors.red} size="large" /><Text style={styles.loading}>LOADING TELEMETRY</Text></View>;
  if (error) return <View style={styles.state}><MaterialCommunityIcons name="access-point-off" size={30} color={colors.red} /><Text style={styles.title}>CONNECTION LOST</Text><Text style={styles.body}>{error.message}{'\n'}{API_URL}</Text>{onRetry ? <Pressable onPress={onRetry} style={styles.retry}><Text style={styles.retryText}>RETRY</Text></Pressable> : null}</View>;
  if (empty) return <View style={styles.state}><MaterialCommunityIcons name="database-off-outline" size={30} color={colors.textMuted} /><Text style={styles.title}>NO DATA SYNCED</Text><Text style={styles.body}>Run the backend synchronization and pull to refresh.</Text></View>;
  return null;
}

const styles = StyleSheet.create({
  state: { minHeight: 240, alignItems: 'center', justifyContent: 'center', padding: spacing.xl, gap: spacing.sm },
  loading: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', letterSpacing: 2, marginTop: spacing.sm },
  title: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 20, letterSpacing: 1 },
  body: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 12, textAlign: 'center', lineHeight: 18 },
  retry: { marginTop: spacing.sm, borderRadius: radius.pill, backgroundColor: colors.red, paddingHorizontal: spacing.xl, paddingVertical: 10 },
  retryText: { color: colors.white, fontFamily: 'Inter_600SemiBold', fontSize: 11, letterSpacing: 1.4 },
});
