import { useMemo } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import Svg, { Circle, Path } from 'react-native-svg';

import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

type Point = { x: number; y: number };

export function TrackMap({ points, sourceLap }: { points: Point[]; sourceLap?: number }) {
  const track = useMemo(() => normalizeTrack(points), [points]);
  if (track.points.length < 2) {
    return <View style={[styles.container, styles.empty]}><Text style={typography.muted}>Circuit location data has not been synchronized.</Text></View>;
  }
  const start = track.points[0];
  return (
    <View style={styles.container}>
      <Svg viewBox="0 0 320 210" width="100%" height={220}>
        <Path d={track.path} fill="none" stroke="#000000" strokeOpacity={0.7} strokeWidth={11} strokeLinejoin="round" strokeLinecap="round" />
        <Path d={track.path} fill="none" stroke={colors.text} strokeOpacity={0.88} strokeWidth={5} strokeLinejoin="round" strokeLinecap="round" />
        <Path d={track.path} fill="none" stroke={colors.red} strokeOpacity={0.28} strokeWidth={2} strokeLinejoin="round" strokeLinecap="round" />
        <Circle cx={start.x} cy={start.y} r={6} fill={colors.background} stroke={colors.red} strokeWidth={3} />
      </Svg>
      <View style={styles.legend}><View style={styles.dot} /><Text style={typography.label}>START / FINISH</Text></View>
      {sourceLap ? <Text style={styles.source}>Generated locally from synchronized lap {sourceLap}</Text> : null}
    </View>
  );
}

function normalizeTrack(points: Point[]) {
  if (points.length < 2) return { points: [] as Point[], path: '' };
  const xs = points.map((point) => point.x);
  const ys = points.map((point) => point.y);
  const minX = Math.min(...xs), maxX = Math.max(...xs), minY = Math.min(...ys), maxY = Math.max(...ys);
  const width = Math.max(maxX - minX, 1), height = Math.max(maxY - minY, 1);
  const scale = Math.min(280 / width, 170 / height);
  const offsetX = (320 - width * scale) / 2;
  const offsetY = (200 - height * scale) / 2;
  const normalized = points.map((point) => ({ x: offsetX + (point.x - minX) * scale, y: offsetY + (maxY - point.y) * scale }));
  return { points: normalized, path: normalized.map((point, index) => `${index ? 'L' : 'M'} ${point.x.toFixed(1)} ${point.y.toFixed(1)}`).join(' ') };
}

const styles = StyleSheet.create({
  container: { minHeight: 250, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, backgroundColor: '#080A0E', overflow: 'hidden', paddingTop: spacing.sm },
  empty: { alignItems: 'center', justifyContent: 'center', padding: spacing.xl },
  legend: { position: 'absolute', left: spacing.md, top: spacing.md, flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  dot: { width: 7, height: 7, borderRadius: 4, backgroundColor: colors.red },
  source: { position: 'absolute', right: spacing.md, bottom: spacing.sm, color: colors.textDim, fontFamily: 'Inter_400Regular', fontSize: 9 },
});
