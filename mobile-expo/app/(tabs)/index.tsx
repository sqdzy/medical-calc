import { View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator } from 'react-native';
import { useRouter } from 'expo-router';
import { useQuery } from '@tanstack/react-query';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { surveysApi, therapyApi } from '../../src/api/client';
import { useAuthStore } from '../../src/store/auth';

export default function HomeScreen() {
  const router = useRouter();
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { data: templates, isLoading: loadingTemplates } = useQuery({
    queryKey: ['survey-templates'],
    queryFn: () => surveysApi.getTemplates(),
    enabled: isAuthenticated,
  });

  const { data: therapyLogs, isLoading: loadingTherapy } = useQuery({
    queryKey: ['therapy-logs', user?.id],
    queryFn: async () => {
      if (!user?.id) return [];
      const res = await therapyApi.listByPatient(user.id);
      return res.data.data;
    },
    enabled: isAuthenticated && !!user?.id,
  });

  if (!isAuthenticated) {
    return (
      <View style={styles.container}>
        <View style={styles.welcomeCard}>
          <FontAwesome name="heartbeat" size={64} color="#2563eb" />
          <Text style={styles.welcomeTitle}>Медицинский калькулятор</Text>
          <Text style={styles.welcomeSubtitle}>
            Система поддержки терапии ГИБП
          </Text>
          <TouchableOpacity
            style={styles.authButton}
            onPress={() => router.push('/auth/login')}
          >
            <Text style={styles.authButtonText}>Войти</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.authButtonSecondary}
            onPress={() => router.push('/auth/register')}
          >
            <Text style={styles.authButtonSecondaryText}>Регистрация</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.greeting}>
          Здравствуйте, {user?.first_name || 'Пациент'}!
        </Text>
        <Text style={styles.date}>
          {new Date().toLocaleDateString('ru-RU', {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
          })}
        </Text>
      </View>

      {/* Quick Actions */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Быстрые действия</Text>
        <View style={styles.actionsRow}>
          <TouchableOpacity
            style={styles.actionCard}
            onPress={() => router.push('/(tabs)/two')}
          >
            <FontAwesome name="list-alt" size={32} color="#2563eb" />
            <Text style={styles.actionText}>Опросники</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.actionCard}
            onPress={() => router.push('/(tabs)/therapy')}
          >
            <FontAwesome name="medkit" size={32} color="#16a34a" />
            <Text style={styles.actionText}>Терапия</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Recent Surveys */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Доступные опросники</Text>
        {loadingTemplates ? (
          <ActivityIndicator size="small" color="#2563eb" />
        ) : templates?.length ? (
          templates.slice(0, 3).map((t: any) => (
            <TouchableOpacity
              key={t.id}
              style={styles.listItem}
              onPress={() => router.push(`/surveys/${t.code}`)}
            >
              <View>
                <Text style={styles.listItemTitle}>{t.name}</Text>
                <Text style={styles.listItemSubtitle}>{t.category}</Text>
              </View>
              <FontAwesome name="chevron-right" size={16} color="#9ca3af" />
            </TouchableOpacity>
          ))
        ) : (
          <Text style={styles.emptyText}>Нет доступных опросников</Text>
        )}
      </View>

      {/* Recent Therapy */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Последняя терапия</Text>
        {loadingTherapy ? (
          <ActivityIndicator size="small" color="#2563eb" />
        ) : therapyLogs?.length ? (
          therapyLogs.slice(0, 3).map((log: any) => (
            <View key={log.id} style={styles.listItem}>
              <View>
                <Text style={styles.listItemTitle}>{log.drug_name}</Text>
                <Text style={styles.listItemSubtitle}>
                  {log.dosage} {log.dosage_unit}
                </Text>
              </View>
              <Text style={styles.dateText}>
                {new Date(log.administered_at).toLocaleDateString('ru-RU')}
              </Text>
            </View>
          ))
        ) : (
          <Text style={styles.emptyText}>Записей о терапии пока нет</Text>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f3f4f6',
  },
  welcomeCard: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  welcomeTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#1f2937',
    marginTop: 24,
    textAlign: 'center',
  },
  welcomeSubtitle: {
    fontSize: 16,
    color: '#6b7280',
    marginTop: 8,
    textAlign: 'center',
  },
  authButton: {
    backgroundColor: '#2563eb',
    paddingHorizontal: 48,
    paddingVertical: 16,
    borderRadius: 12,
    marginTop: 32,
    width: '100%',
  },
  authButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
    textAlign: 'center',
  },
  authButtonSecondary: {
    paddingHorizontal: 48,
    paddingVertical: 16,
    marginTop: 12,
  },
  authButtonSecondaryText: {
    color: '#2563eb',
    fontSize: 16,
    fontWeight: '500',
    textAlign: 'center',
  },
  header: {
    backgroundColor: '#2563eb',
    padding: 24,
    paddingTop: 60,
  },
  greeting: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  date: {
    fontSize: 14,
    color: '#bfdbfe',
    marginTop: 4,
  },
  section: {
    padding: 16,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#374151',
    marginBottom: 12,
  },
  actionsRow: {
    flexDirection: 'row',
    gap: 12,
  },
  actionCard: {
    flex: 1,
    backgroundColor: '#fff',
    padding: 20,
    borderRadius: 12,
    alignItems: 'center',
    shadowColor: '#000',
    shadowOpacity: 0.05,
    shadowRadius: 4,
    shadowOffset: { width: 0, height: 2 },
    elevation: 2,
  },
  actionText: {
    marginTop: 8,
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
  },
  listItem: {
    backgroundColor: '#fff',
    padding: 16,
    borderRadius: 12,
    marginBottom: 8,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  listItemTitle: {
    fontSize: 16,
    fontWeight: '500',
    color: '#1f2937',
  },
  listItemSubtitle: {
    fontSize: 14,
    color: '#6b7280',
    marginTop: 2,
  },
  dateText: {
    fontSize: 12,
    color: '#9ca3af',
  },
  emptyText: {
    textAlign: 'center',
    color: '#9ca3af',
    padding: 16,
  },
});
