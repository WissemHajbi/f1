import { StyleSheet } from 'react-native';

import { colors } from './colors';

export const typography = StyleSheet.create({
  body: { color: colors.text, fontFamily: 'Inter_400Regular', fontSize: 14 },
  muted: { color: colors.textMuted, fontFamily: 'Inter_400Regular', fontSize: 12 },
  label: { color: colors.textMuted, fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 12, letterSpacing: 1.5 },
  value: { color: colors.text, fontFamily: 'BarlowCondensed_700Bold', fontSize: 22 },
});
