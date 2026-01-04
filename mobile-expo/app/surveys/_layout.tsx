import { Stack } from 'expo-router';

export default function SurveysLayout() {
  return (
    <Stack
      screenOptions={{
        headerStyle: { backgroundColor: '#2563eb' },
        headerTintColor: '#fff',
        headerTitleStyle: { fontWeight: '600' },
      }}
    >
      <Stack.Screen
        name="[code]"
        options={{
          title: 'Опросник',
        }}
      />
    </Stack>
  );
}
