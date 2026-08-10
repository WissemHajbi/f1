import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMemo, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Gesture, GestureDetector } from 'react-native-gesture-handler';
import Animated, { useAnimatedStyle, useSharedValue } from 'react-native-reanimated';
import Svg, { Circle, Path } from 'react-native-svg';

import { colors } from '@/theme/colors';
import { radius, spacing } from '@/theme/spacing';
import { typography } from '@/theme/typography';

type Point = { x: number; y: number; timestamp?: string };
type ProjectedPoint = Point & { rawX: number; rawY: number };

export function TrackMap({ points, trace = [], traceColor = colors.red, sourceLap, traceLap, attribution, accuracy, estimatedWidth }: {
  points: Point[]; trace?: Point[]; traceColor?: string; sourceLap?: number; traceLap?: number;
  attribution?: string; accuracy?: string; estimatedWidth?: number;
}) {
  const track = useMemo(() => projectTrack(points, trace), [points, trace]);
  const [selectedPoint, setSelectedPoint] = useState<number>();
  const zoom = useSharedValue(1), savedZoom = useSharedValue(1);
  const offsetX = useSharedValue(0), offsetY = useSharedValue(0), savedX = useSharedValue(0), savedY = useSharedValue(0);
  const pinch = Gesture.Pinch().onUpdate((event) => { zoom.value = Math.max(1, Math.min(8, savedZoom.value * event.scale)); })
    .onEnd(() => { savedZoom.value = zoom.value; });
  const pan = Gesture.Pan().onUpdate((event) => { offsetX.value = savedX.value + event.translationX; offsetY.value = savedY.value + event.translationY; })
    .onEnd(() => { savedX.value = offsetX.value; savedY.value = offsetY.value; });
  const style = useAnimatedStyle(() => ({ transform: [{ translateX: offsetX.value }, { translateY: offsetY.value }, { scale: zoom.value }] }));
  const reset = () => { zoom.value = 1; savedZoom.value = 1; offsetX.value = 0; offsetY.value = 0; savedX.value = 0; savedY.value = 0; setSelectedPoint(undefined); };
  const inspected = selectedPoint === undefined ? undefined : track.trace[selectedPoint];

  if (track.points.length < 2) return <View style={[styles.container, styles.empty]}><Text style={typography.muted}>Circuit geometry is not synchronized.</Text></View>;

  return <View style={styles.container}>
    <View style={styles.header}>
      <View><Text style={typography.label}>{traceLap ? `DRIVER LAP ${traceLap}` : 'CIRCUIT MAP'}</Text><Text style={styles.hint}>Pinch to zoom · drag to pan · tap a sample</Text></View>
      <Pressable accessibilityRole="button" accessibilityLabel="Reset circuit view" onPress={reset} style={styles.reset}><MaterialCommunityIcons name="fit-to-screen-outline" size={20} color={colors.text} /></Pressable>
    </View>
    <View style={styles.viewport}>
      <GestureDetector gesture={Gesture.Simultaneous(pinch, pan)}>
        <Animated.View style={[styles.canvas, style]}>
          <Svg viewBox="0 0 360 380" width="100%" height="100%">
            <Path d={track.trackPath} fill="none" stroke="#020305" strokeWidth={30} strokeLinejoin="round" strokeLinecap="round" />
            <Path d={track.trackPath} fill="none" stroke={colors.text} strokeOpacity={0.92} strokeWidth={25} strokeLinejoin="round" strokeLinecap="round" />
            <Path d={track.trackPath} fill="none" stroke="#10141B" strokeWidth={18} strokeLinejoin="round" strokeLinecap="round" />
            {track.tracePath ? <Path d={track.tracePath} fill="none" stroke="#000" strokeOpacity={0.85} strokeWidth={4.4} strokeLinejoin="round" strokeLinecap="round" /> : null}
            {track.tracePath ? <Path d={track.tracePath} fill="none" stroke={traceColor} strokeWidth={2.1} strokeLinejoin="round" strokeLinecap="round" /> : null}
            {track.trace.map((point, index) => <Circle key={index} cx={point.x} cy={point.y}
              r={selectedPoint === index ? 4 : 1.15} fill={selectedPoint === index ? colors.text : traceColor}
              stroke={selectedPoint === index ? traceColor : 'transparent'} strokeWidth={1.5} onPress={() => setSelectedPoint(index)} />)}
          </Svg>
        </Animated.View>
      </GestureDetector>
      {!track.trace.length ? <Text style={styles.noTrace}>{"This lap's location window is not synchronized."}</Text> : null}
    </View>
    {inspected ? <View style={styles.inspector}>
      <View><Text style={typography.label}>SAMPLE {selectedPoint! + 1} / {track.trace.length}</Text><Text style={styles.coordinates}>X {inspected.rawX} · Y {inspected.rawY}</Text></View>
      <Text style={styles.time}>{formatTime(inspected.timestamp)}</Text>
    </View> : null}
    <View style={styles.metadata}>
      <Text style={styles.metaText}>{estimatedWidth ? `Visual width ≈ ${estimatedWidth} m · ` : ''}{accuracy || (sourceLap ? `Trajectory reference lap ${sourceLap}` : 'Estimated geometry')}</Text>
      {attribution ? <Text style={styles.attribution}>{attribution}</Text> : null}
    </View>
  </View>;
}

function projectTrack(track: Point[], trace: Point[]) {
  if (track.length < 2) return { points: [] as ProjectedPoint[], trace: [] as ProjectedPoint[], trackPath: '', tracePath: '' };
  const minX = Math.min(...track.map((p) => p.x)), maxX = Math.max(...track.map((p) => p.x));
  const minY = Math.min(...track.map((p) => p.y)), maxY = Math.max(...track.map((p) => p.y));
  const width = Math.max(maxX-minX, 1), height = Math.max(maxY-minY, 1), scale = Math.min(315/width, 325/height);
  const left = (360-width*scale)/2, top = (380-height*scale)/2;
  const project = (values: Point[]): ProjectedPoint[] => values.map((point) => ({ ...point, rawX: point.x, rawY: point.y, x: left+(point.x-minX)*scale, y: top+(maxY-point.y)*scale }));
  const projectedTrack = project(track), projectedTrace = project(trace);
  return { points: projectedTrack, trace: projectedTrace, trackPath: path(projectedTrack), tracePath: path(projectedTrace) };
}
function path(points: Point[]) { return points.map((point, index) => `${index ? 'L' : 'M'} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`).join(' '); }
function formatTime(timestamp?: string) { return timestamp ? new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 2 }) : 'Time unavailable'; }

const styles = StyleSheet.create({
  container: { minHeight: 570, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, backgroundColor: '#080A0E', overflow: 'hidden' },
  empty: { alignItems: 'center', justifyContent: 'center', padding: spacing.xl },
  header: { height: 62, paddingHorizontal: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', borderBottomWidth: 1, borderBottomColor: colors.border, backgroundColor: colors.surface },
  hint: { color: colors.textDim, fontFamily: 'Inter_400Regular', fontSize: 9, marginTop: 3 },
  reset: { width: 38, height: 38, borderRadius: radius.pill, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.surfaceRaised },
  viewport: { height: 420, overflow: 'hidden' },
  canvas: { width: '100%', height: 420 },
  noTrace: { position: 'absolute', top: 195, left: spacing.md, right: spacing.md, textAlign: 'center', color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 11 },
  inspector: { minHeight: 60, paddingHorizontal: spacing.md, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', borderTopWidth: 1, borderTopColor: colors.border, backgroundColor: colors.surface },
  coordinates: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 17, marginTop: 2 },
  time: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 11 },
  metadata: { minHeight: 47, paddingHorizontal: spacing.md, justifyContent: 'center', borderTopWidth: 1, borderTopColor: colors.border },
  metaText: { color: colors.textDim, fontFamily: 'Inter_400Regular', fontSize: 8 },
  attribution: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 9, marginTop: 2 },
});
