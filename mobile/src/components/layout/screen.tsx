import { LinearGradient } from 'expo-linear-gradient';
import type { PropsWithChildren } from 'react';
import { RefreshControl, ScrollView, StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { colors } from '@/theme/colors';
import { spacing } from '@/theme/spacing';

export function Screen({ children, refreshing, onRefresh }: PropsWithChildren<{ refreshing?: boolean; onRefresh?: () => void }>) {
  return (
    <SafeAreaView style={styles.safe} edges={['top']}>
      <View pointerEvents="none" style={StyleSheet.absoluteFill}>
        <LinearGradient colors={['#17060B', colors.background, colors.background]} locations={[0, 0.28, 1]} style={StyleSheet.absoluteFill} />
        <View style={styles.speedLineOne} /><View style={styles.speedLineTwo} />
      </View>
      <ScrollView contentContainerStyle={styles.content} showsVerticalScrollIndicator={false}
        refreshControl={onRefresh ? <RefreshControl tintColor={colors.red} refreshing={Boolean(refreshing)} onRefresh={onRefresh} /> : undefined}>
        {children}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  content: { paddingHorizontal: spacing.lg, paddingBottom: 120, gap: spacing.xl },
  speedLineOne: { position: 'absolute', width: 260, height: 1, backgroundColor: '#60101F', top: 80, right: -60, transform: [{ rotate: '-22deg' }] },
  speedLineTwo: { position: 'absolute', width: 180, height: 1, backgroundColor: '#301019', top: 116, right: -40, transform: [{ rotate: '-22deg' }] },
});
