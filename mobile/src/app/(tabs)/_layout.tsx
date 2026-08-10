import { MaterialCommunityIcons } from '@expo/vector-icons';
import { Tabs } from 'expo-router';
import type { ColorValue } from 'react-native';

import { colors } from '@/theme/colors';

type IconName = keyof typeof MaterialCommunityIcons.glyphMap;
const icon = (name: IconName) => {
  const TabIcon = ({ color, size }: { color: ColorValue; size: number }) => <MaterialCommunityIcons name={name} color={color} size={size} />;
  TabIcon.displayName = `${name}TabIcon`;
  return TabIcon;
};

export default function TabsLayout() {
  return (
    <Tabs screenOptions={{
      headerShown: false,
      tabBarActiveTintColor: colors.red,
      tabBarInactiveTintColor: colors.textDim,
      tabBarStyle: { position: 'absolute', height: 76, paddingTop: 8, paddingBottom: 12, backgroundColor: '#0D0F13F5', borderTopColor: colors.border },
      tabBarLabelStyle: { fontFamily: 'BarlowCondensed_600SemiBold', fontSize: 11, letterSpacing: 0.8, textTransform: 'uppercase' },
      sceneStyle: { backgroundColor: colors.background },
    }}>
      <Tabs.Screen name="index" options={{ title: 'Home', tabBarIcon: icon('flag-checkered') }} />
      <Tabs.Screen name="calendar" options={{ title: 'Calendar', tabBarIcon: icon('calendar-blank-outline') }} />
      <Tabs.Screen name="standings" options={{ title: 'Standings', tabBarIcon: icon('podium') }} />
      <Tabs.Screen name="drivers" options={{ title: 'Drivers', tabBarIcon: icon('account-group-outline') }} />
    </Tabs>
  );
}
