import { StyleSheet, Text, View } from 'react-native';
import Svg, { Circle, Path } from 'react-native-svg';

import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

type Point = { x: number; y: number; timestamp?: string };

export function TrackMap({ points, trace = [], traceColor = colors.red, sourceLap, traceLap }: {
  points: Point[];
  trace?: Point[];
  traceColor?: string;
  sourceLap?: number;
  traceLap?: number;
}) {
  const track = projectTrack(points, trace);
  if (track.points.length < 2) {
    return <View style={[styles.container, styles.empty]}><Text style={typography.muted}>Circuit location data has not been synchronized.</Text></View>;
  }
  return <View style={styles.container}>
    <View style={styles.header}><Text style={typography.label}>{traceLap ? `DRIVER LAP ${traceLap}` : 'CIRCUIT REFERENCE'}</Text></View>
    <Svg viewBox="0 0 340 350" width="100%" height={400}>
      <Path d={track.path} fill="none" stroke="#020305" strokeWidth={28} strokeLinejoin="round" strokeLinecap="round" />
      <Path d={track.path} fill="none" stroke={colors.text} strokeOpacity={0.9} strokeWidth={23} strokeLinejoin="round" strokeLinecap="round" />
      <Path d={track.path} fill="none" stroke="#10141B" strokeWidth={17} strokeLinejoin="round" strokeLinecap="round" />
      {track.tracePath ? <Path d={track.tracePath} fill="none" stroke={traceColor} strokeWidth={2} strokeLinejoin="round" strokeLinecap="round" /> : null}
      {track.trace.map((point, index) => <Circle key={index} cx={point.x} cy={point.y} r={1.1} fill={traceColor} />)}
    </Svg>
    {!track.trace.length ? <Text style={styles.noTrace}>Location samples for this lap are not synchronized.</Text> : null}
    {sourceLap ? <Text style={styles.source}>Inferred corridor around reference lap {sourceLap}</Text> : null}
  </View>;
}

function projectTrack(points: Point[], trace: Point[]) {
  if (points.length < 2) return { points: [] as Point[], trace: [] as Point[], path: '', tracePath: '' };
  const xs = points.map((point) => point.x), ys = points.map((point) => point.y);
  const minX = Math.min(...xs), maxX = Math.max(...xs), minY = Math.min(...ys), maxY = Math.max(...ys);
  const width = Math.max(maxX - minX, 1), height = Math.max(maxY - minY, 1);
  const scale = Math.min(285 / width, 285 / height);
  const offsetX = (340 - width * scale) / 2, offsetY = (350 - height * scale) / 2;
  const project = (values: Point[]) => values.map((point) => ({ ...point, x: offsetX + (point.x - minX) * scale, y: offsetY + (maxY - point.y) * scale }));
  const normalized = project(points), normalizedTrace = project(trace);
  const path = (values: Point[]) => values.map((point, index) => `${index ? 'L' : 'M'} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`).join(' ');
  return { points: normalized, trace: normalizedTrace, path: path(normalized), tracePath: path(normalizedTrace) };
}

const styles = StyleSheet.create({
  container: { minHeight: 470, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, backgroundColor: '#080A0E', overflow: 'hidden' },
  empty: { alignItems: 'center', justifyContent: 'center', padding: spacing.xl },
  header: { minHeight: 48, justifyContent: 'center', paddingHorizontal: spacing.md, borderBottomWidth: 1, borderBottomColor: colors.border },
  noTrace: { position: 'absolute', top: 240, left: spacing.md, right: spacing.md, color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 11, textAlign: 'center' },
  source: { position: 'absolute', bottom: spacing.xs, right: spacing.sm, color: colors.textDim, fontFamily: 'Inter_400Regular', fontSize: 8 },
});
