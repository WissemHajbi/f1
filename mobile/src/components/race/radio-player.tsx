import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { useAudioPlayer, useAudioPlayerStatus } from 'expo-audio';
import { useEffect, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { API_URL } from '@/api/client';
import { queries } from '@/api/hooks';
import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

export function RadioPlayer({ sessionKey, driverNumber, count, color }: {
  sessionKey: number;
  driverNumber: number;
  count: number;
  color: string;
}) {
  const query = useQuery(queries.teamRadio(sessionKey, driverNumber, count > 0));
  const player = useAudioPlayer(undefined, { updateInterval: 250 });
  const status = useAudioPlayerStatus(player);
  const [active, setActive] = useState<string>();

  useEffect(() => {
    player.pause();
    setActive(undefined);
  }, [driverNumber, player]);

  const toggle = (id: string, audioURL: string) => {
    if (active === id) {
      if (status.playing) player.pause(); else player.play();
      return;
    }
    player.replace(`${API_URL}${audioURL}`);
    setActive(id);
    player.play();
  };

  if (count === 0) return <Text style={typography.muted}>No synchronized radio recordings for this driver.</Text>;
  if (query.isLoading) return <Text style={typography.muted}>Loading local radio recordings…</Text>;
  if (query.error) return <Text style={typography.muted}>Radio recordings could not be loaded.</Text>;

  return <View style={styles.list}>{query.data?.map((radio, index) => {
    const playing = active === radio.id && status.playing;
    const progress = active === radio.id && status.duration > 0 ? status.currentTime / status.duration : 0;
    return <Pressable key={radio.id} accessibilityRole="button" accessibilityLabel={`${playing ? 'Pause' : 'Play'} radio ${index + 1}`}
      onPress={() => toggle(radio.id, radio.audio_url)} style={[styles.row, active === radio.id && { borderColor: color }]}>
      <View style={[styles.play, { backgroundColor: color }]}><MaterialCommunityIcons name={playing ? 'pause' : 'play'} color={colors.background} size={19} /></View>
      <View style={styles.copy}>
        <Text style={styles.title}>RADIO {String(index + 1).padStart(2, '0')}</Text>
        <Text style={typography.muted}>{new Date(radio.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })}</Text>
        <View style={styles.progress}><View style={[styles.progressFill, { width: `${Math.min(progress * 100, 100)}%`, backgroundColor: color }]} /></View>
      </View>
      {active === radio.id && status.duration > 0 ? <Text style={styles.time}>{Math.floor(status.currentTime)}s</Text> : null}
    </Pressable>;
  })}</View>;
}

const styles = StyleSheet.create({
  list: { gap: spacing.sm },
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, padding: spacing.sm, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surfaceRaised },
  play: { width: 38, height: 38, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center' },
  copy: { flex: 1 },
  title: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 14, letterSpacing: 1 },
  progress: { height: 3, borderRadius: radius.pill, backgroundColor: colors.border, marginTop: spacing.xs, overflow: 'hidden' },
  progressFill: { height: '100%', borderRadius: radius.pill },
  time: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 12 },
});
