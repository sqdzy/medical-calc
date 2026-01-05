import { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  TextInput,
  Modal,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import FontAwesome from '@expo/vector-icons/FontAwesome';
import { therapyApi, drugsApi } from '../../src/api/client';
import { useAuthStore } from '../../src/store/auth';

interface TherapyLog {
  id: string;
  drug_id: string;
  drug_name: string;
  dosage: string;
  dosage_unit: string;
  route: string;
  administered_at?: string;
  next_scheduled?: string;
  notes?: string;
}

function formatRuDate(value?: string) {
  if (!value) return '—';
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  });
}

function normalizeDosageInput(value: string) {
  // allow digits and one decimal separator ('.' or ',')
  const cleaned = value.replace(/[^0-9.,]/g, '');
  const firstSepIndex = cleaned.search(/[.,]/);
  if (firstSepIndex === -1) return cleaned;
  const head = cleaned.slice(0, firstSepIndex + 1);
  const tail = cleaned.slice(firstSepIndex + 1).replace(/[.,]/g, '');
  return head + tail;
}

interface Drug {
  id: string;
  name: string;
  international_name: string;
  category: string;
  standard_dosage?: number;
  dosage_unit?: string;
}

export default function TherapyScreen() {
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const queryClient = useQueryClient();

  const [showAddModal, setShowAddModal] = useState(false);
  const [selectedDrug, setSelectedDrug] = useState<Drug | null>(null);
  const [dosage, setDosage] = useState('');
  const [dosageUnit, setDosageUnit] = useState('мг');
  const [notes, setNotes] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // Fetch therapy logs
  const { data: therapyLogs, isLoading: loadingLogs, refetch } = useQuery({
    queryKey: ['therapy-logs', user?.id],
    queryFn: async () => {
      if (!user?.id) return [];
      const res = await therapyApi.listByPatient(user.id);
      return res.data.data || [];
    },
    enabled: isAuthenticated && !!user?.id,
  });

  // Fetch drugs list
  const { data: drugs, isLoading: loadingDrugs } = useQuery({
    queryKey: ['drugs', searchQuery],
    queryFn: async () => {
      const res = await drugsApi.list(searchQuery || undefined);
      return res.data.data || [];
    },
    enabled: showAddModal,
  });

  // Create therapy log mutation
  const createLogMutation = useMutation({
    mutationFn: async (data: {
      patient_id: string;
      drug_id: string;
      dosage: string;
      dosage_unit: string;
      notes?: string;
    }) => {
      return therapyApi.createLog(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['therapy-logs', user?.id] });
      setShowAddModal(false);
      resetForm();
      Alert.alert('Успешно', 'Запись о терапии добавлена');
    },
    onError: (error: any) => {
      Alert.alert('Ошибка', error.response?.data?.message || 'Не удалось добавить запись');
    },
  });

  const deleteLogMutation = useMutation({
    mutationFn: async (logId: string) => {
      return therapyApi.deleteLog(logId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['therapy-logs', user?.id] });
      Alert.alert('Успешно', 'Запись удалена');
    },
    onError: (error: any) => {
      Alert.alert('Ошибка', error.response?.data?.message || 'Не удалось удалить запись');
    },
  });

  const resetForm = () => {
    setSelectedDrug(null);
    setDosage('');
    setDosageUnit('мг');
    setNotes('');
    setSearchQuery('');
  };

  const handleAddLog = () => {
    if (!selectedDrug || !dosage || !user?.id) {
      Alert.alert('Ошибка', 'Выберите препарат и укажите дозировку');
      return;
    }

    const normalizedDosage = dosage.replace(',', '.').trim();

    const parsed = Number.parseFloat(normalizedDosage);
    if (Number.isNaN(parsed) || parsed <= 0) {
      Alert.alert('Ошибка', 'Дозировка должна быть числом больше 0');
      return;
    }

    createLogMutation.mutate({
      patient_id: user.id,
      drug_id: selectedDrug.id,
      dosage: normalizedDosage,
      dosage_unit: dosageUnit,
      notes: notes || undefined,
    });
  };

  if (!isAuthenticated) {
    return (
      <View style={styles.centerContainer}>
        <FontAwesome name="lock" size={64} color="#9ca3af" />
        <Text style={styles.emptyTitle}>Требуется авторизация</Text>
        <Text style={styles.emptyText}>Войдите в аккаунт для доступа к терапии</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Терапия ГИБП</Text>
        <TouchableOpacity
          style={styles.addButton}
          onPress={() => setShowAddModal(true)}
        >
          <FontAwesome name="plus" size={20} color="#fff" />
        </TouchableOpacity>
      </View>

      {/* Content */}
      <ScrollView style={styles.content}>
        {loadingLogs ? (
          <View style={styles.centerContainer}>
            <ActivityIndicator size="large" color="#2563eb" />
          </View>
        ) : therapyLogs?.length ? (
          therapyLogs.map((log: TherapyLog) => (
            <View key={log.id} style={styles.logCard}>
              <View style={styles.logHeader}>
                <View style={styles.logHeaderLeft}>
                  <FontAwesome name="medkit" size={20} color="#16a34a" />
                  <Text style={styles.logDrugName}>{log.drug_name}</Text>
                </View>
                <TouchableOpacity
                  onPress={() => {
                    Alert.alert('Удалить запись?', 'Это действие нельзя отменить.', [
                      { text: 'Отмена', style: 'cancel' },
                      {
                        text: 'Удалить',
                        style: 'destructive',
                        onPress: () => deleteLogMutation.mutate(log.id),
                      },
                    ]);
                  }}
                  disabled={deleteLogMutation.isPending}
                  accessibilityLabel="Удалить запись"
                >
                  <FontAwesome name="trash" size={18} color="#dc2626" />
                </TouchableOpacity>
              </View>
              <View style={styles.logDetails}>
                <View style={styles.logDetailRow}>
                  <Text style={styles.logLabel}>Дозировка:</Text>
                  <Text style={styles.logValue}>{log.dosage} {log.dosage_unit}</Text>
                </View>
                <View style={styles.logDetailRow}>
                  <Text style={styles.logLabel}>Дата введения:</Text>
                  <Text style={styles.logValue}>
                    {formatRuDate(log.administered_at)}
                  </Text>
                </View>
                {log.next_scheduled && (
                  <View style={styles.logDetailRow}>
                    <Text style={styles.logLabel}>Следующее введение:</Text>
                    <Text style={styles.logValue}>
                      {formatRuDate(log.next_scheduled)}
                    </Text>
                  </View>
                )}
                {log.notes && (
                  <View style={styles.logNotes}>
                    <Text style={styles.logNotesText}>{log.notes}</Text>
                  </View>
                )}
              </View>
            </View>
          ))
        ) : (
          <View style={styles.emptyContainer}>
            <FontAwesome name="calendar-o" size={64} color="#9ca3af" />
            <Text style={styles.emptyTitle}>Записей пока нет</Text>
            <Text style={styles.emptyText}>
              Нажмите + чтобы добавить первую запись о терапии
            </Text>
          </View>
        )}
      </ScrollView>

      {/* Add Therapy Modal */}
      <Modal
        visible={showAddModal}
        animationType="slide"
        presentationStyle="pageSheet"
        onRequestClose={() => {
          setShowAddModal(false);
          resetForm();
        }}
      >
        <SafeAreaView style={styles.modalContainer}>
          <View style={styles.modalHeader}>
            <TouchableOpacity
              onPress={() => {
                setShowAddModal(false);
                resetForm();
              }}
            >
              <Text style={styles.modalCancel}>Отмена</Text>
            </TouchableOpacity>
            <Text style={styles.modalTitle}>Новая запись</Text>
            <TouchableOpacity
              onPress={handleAddLog}
              disabled={createLogMutation.isPending}
            >
              <Text style={[styles.modalSave, createLogMutation.isPending && styles.modalSaveDisabled]}>
                {createLogMutation.isPending ? 'Сохр...' : 'Сохранить'}
              </Text>
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.modalContent}>
            {/* Drug Selection */}
            <Text style={styles.inputLabel}>Препарат</Text>
            {selectedDrug ? (
              <TouchableOpacity
                style={styles.selectedDrug}
                onPress={() => setSelectedDrug(null)}
              >
                <View>
                  <Text style={styles.selectedDrugName}>{selectedDrug.name}</Text>
                  <Text style={styles.selectedDrugSub}>{selectedDrug.international_name}</Text>
                </View>
                <FontAwesome name="times" size={20} color="#6b7280" />
              </TouchableOpacity>
            ) : (
              <>
                <TextInput
                  style={styles.searchInput}
                  placeholder="Поиск препарата..."
                  placeholderTextColor="#9ca3af"
                  value={searchQuery}
                  onChangeText={setSearchQuery}
                />
                {loadingDrugs ? (
                  <ActivityIndicator size="small" color="#2563eb" style={styles.loader} />
                ) : drugs?.length ? (
                  <View style={styles.drugsList}>
                    {drugs.slice(0, 10).map((drug: Drug) => (
                      <TouchableOpacity
                        key={drug.id}
                        style={styles.drugItem}
                        onPress={() => {
                          setSelectedDrug(drug);
                          if (drug.standard_dosage) {
                            setDosage(drug.standard_dosage.toString());
                          }
                          if (drug.dosage_unit) {
                            setDosageUnit(drug.dosage_unit);
                          }
                        }}
                      >
                        <Text style={styles.drugName}>{drug.name}</Text>
                        <Text style={styles.drugCategory}>{drug.category}</Text>
                      </TouchableOpacity>
                    ))}
                  </View>
                ) : searchQuery ? (
                  <Text style={styles.noResults}>Препараты не найдены</Text>
                ) : null}
              </>
            )}

            {/* Dosage */}
            <Text style={styles.inputLabel}>Дозировка</Text>
            <View style={styles.dosageRow}>
              <TextInput
                style={[styles.input, styles.dosageInput]}
                placeholder="100"
                placeholderTextColor="#9ca3af"
                keyboardType="decimal-pad"
                value={dosage}
                onChangeText={(v) => setDosage(normalizeDosageInput(v))}
              />
              <View style={styles.unitPicker}>
                {['мг', 'мл', 'ЕД'].map((unit) => (
                  <TouchableOpacity
                    key={unit}
                    style={[styles.unitButton, dosageUnit === unit && styles.unitButtonActive]}
                    onPress={() => setDosageUnit(unit)}
                  >
                    <Text style={[styles.unitButtonText, dosageUnit === unit && styles.unitButtonTextActive]}>
                      {unit}
                    </Text>
                  </TouchableOpacity>
                ))}
              </View>
            </View>

            {/* Notes */}
            <Text style={styles.inputLabel}>Заметки (опционально)</Text>
            <TextInput
              style={[styles.input, styles.notesInput]}
              placeholder="Дополнительная информация..."
              placeholderTextColor="#9ca3af"
              multiline
              numberOfLines={3}
              value={notes}
              onChangeText={setNotes}
            />
          </ScrollView>
        </SafeAreaView>
      </Modal>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f3f4f6',
  },
  header: {
    backgroundColor: '#2563eb',
    paddingTop: 60,
    paddingBottom: 20,
    paddingHorizontal: 20,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  addButton: {
    backgroundColor: 'rgba(255,255,255,0.2)',
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    flex: 1,
    padding: 16,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  logCard: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
  },
  logHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 12,
    paddingBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#f3f4f6',
  },
  logHeaderLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    paddingRight: 12,
  },
  logDrugName: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1f2937',
    marginLeft: 12,
  },
  logDetails: {},
  logDetailRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  logLabel: {
    fontSize: 14,
    color: '#6b7280',
  },
  logValue: {
    fontSize: 14,
    fontWeight: '500',
    color: '#1f2937',
  },
  logNotes: {
    marginTop: 8,
    padding: 12,
    backgroundColor: '#f9fafb',
    borderRadius: 8,
  },
  logNotesText: {
    fontSize: 14,
    color: '#6b7280',
    fontStyle: 'italic',
  },
  emptyContainer: {
    alignItems: 'center',
    paddingTop: 80,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#1f2937',
    marginTop: 16,
  },
  emptyText: {
    fontSize: 14,
    color: '#6b7280',
    marginTop: 8,
    textAlign: 'center',
  },
  modalContainer: {
    flex: 1,
    backgroundColor: '#fff',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1f2937',
  },
  modalCancel: {
    fontSize: 16,
    color: '#6b7280',
  },
  modalSave: {
    fontSize: 16,
    fontWeight: '600',
    color: '#2563eb',
  },
  modalSaveDisabled: {
    color: '#9ca3af',
  },
  modalContent: {
    flex: 1,
    padding: 16,
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    marginBottom: 8,
    marginTop: 16,
  },
  input: {
    backgroundColor: '#f3f4f6',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  searchInput: {
    backgroundColor: '#f3f4f6',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  selectedDrug: {
    backgroundColor: '#eff6ff',
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2563eb',
  },
  selectedDrugName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#1f2937',
  },
  selectedDrugSub: {
    fontSize: 14,
    color: '#6b7280',
    marginTop: 2,
  },
  loader: {
    marginTop: 16,
  },
  drugsList: {
    marginTop: 8,
  },
  drugItem: {
    padding: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  drugName: {
    fontSize: 16,
    color: '#1f2937',
  },
  drugCategory: {
    fontSize: 12,
    color: '#6b7280',
    marginTop: 2,
  },
  noResults: {
    textAlign: 'center',
    color: '#6b7280',
    marginTop: 16,
  },
  dosageRow: {
    flexDirection: 'row',
    gap: 12,
  },
  dosageInput: {
    flex: 1,
  },
  unitPicker: {
    flexDirection: 'row',
    gap: 8,
  },
  unitButton: {
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderRadius: 12,
    backgroundColor: '#f3f4f6',
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  unitButtonActive: {
    backgroundColor: '#2563eb',
    borderColor: '#2563eb',
  },
  unitButtonText: {
    fontSize: 16,
    color: '#6b7280',
  },
  unitButtonTextActive: {
    color: '#fff',
    fontWeight: '500',
  },
  notesInput: {
    height: 100,
    textAlignVertical: 'top',
  },
});
