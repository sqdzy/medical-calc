-- 005_preop_risk_surveys.up.sql
-- Replace old survey templates with preoperative risk assessment scales

-- First, clear existing templates and add new ones for preoperative risk assessment
DELETE FROM survey_templates WHERE code IN ('BVAS_V3', 'DAS28_CRP', 'BASDAI');

-- ASA Physical Status Classification
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000001',
    'ASA',
    'ASA Physical Status Classification',
    'Классификация физического статуса ASA (American Society of Anesthesiologists) для оценки предоперационного состояния пациента',
    'preoperative',
    $$[
        {
            "section": "classification",
            "title": "Выберите класс ASA",
            "questions": [
                {
                    "id": "asa_class",
                    "text": "Физический статус пациента",
                    "type": "select",
                    "options": [
                        {"value": 1, "label": "ASA I - Здоровый пациент", "description": "Нет органических, физиологических или психических нарушений."},
                        {"value": 2, "label": "ASA II - Лёгкое системное заболевание", "description": "Контролируемая гипертензия, диабет без осложнений, ожирение (ИМТ 30-40), курение."},
                        {"value": 3, "label": "ASA III - Тяжёлое системное заболевание", "description": "Плохо контролируемый диабет/гипертензия, ХОБЛ, морбидное ожирение, ХПН на диализе, ИБС, ХСН."},
                        {"value": 4, "label": "ASA IV - Угрожающее жизни заболевание", "description": "Недавний ИМ (<3 мес), инсульт, ТИА, тяжёлый сепсис, ДВС, ОРДС."},
                        {"value": 5, "label": "ASA V - Умирающий пациент", "description": "Разрыв аневризмы аорты, массивная травма, внутричерепное кровоизлияние."},
                        {"value": 6, "label": "ASA VI - Донор органов", "description": "Пациент с диагностированной смертью мозга."}
                    ]
                },
                {
                    "id": "is_emergency",
                    "text": "Экстренная операция?",
                    "type": "boolean",
                    "description": "Добавляет модификатор E к классу ASA. Экстренная операция - когда задержка лечения увеличивает угрозу жизни."
                }
            ]
        },
        {
            "section": "details",
            "title": "Дополнительная информация",
            "questions": [
                {
                    "id": "age",
                    "text": "Возраст пациента (лет)",
                    "type": "number",
                    "min": 0,
                    "max": 120
                },
                {
                    "id": "surgery_type",
                    "text": "Тип планируемой операции",
                    "type": "text"
                },
                {
                    "id": "comorbidities",
                    "text": "Основные сопутствующие заболевания",
                    "type": "text"
                }
            ]
        }
    ]$$::jsonb,
    $${
        "type": "direct",
        "field": "asa_class"
    }$$::jsonb,
    $${
        "ranges": [
            {"min": 1, "max": 1, "category": "ASA I", "description": "Здоровый пациент. Периоперационная смертность ~0.1%"},
            {"min": 2, "max": 2, "category": "ASA II", "description": "Лёгкое системное заболевание. Периоперационная смертность ~0.2%"},
            {"min": 3, "max": 3, "category": "ASA III", "description": "Тяжёлое системное заболевание. Периоперационная смертность ~1.8%"},
            {"min": 4, "max": 4, "category": "ASA IV", "description": "Угрожающее жизни заболевание. Периоперационная смертность ~7.8%"},
            {"min": 5, "max": 5, "category": "ASA V", "description": "Умирающий пациент. Периоперационная смертность ~9.4%"},
            {"min": 6, "max": 6, "category": "ASA VI", "description": "Донор органов (смерть мозга)"}
        ],
        "emergency_modifier": "При экстренной операции (+E) риски увеличиваются"
    }$$::jsonb,
    1
)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    questions = EXCLUDED.questions,
    scoring_logic = EXCLUDED.scoring_logic,
    interpretation_rules = EXCLUDED.interpretation_rules,
    version = EXCLUDED.version,
    updated_at = NOW();

-- Lee Revised Cardiac Risk Index (RCRI)
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000002',
    'RCRI',
    'Revised Cardiac Risk Index (Lee)',
    'Пересмотренный индекс кардиального риска (RCRI, индекс Lee) для оценки риска сердечных осложнений при некардиохирургических операциях',
    'preoperative',
    $$[
        {
            "section": "risk_factors",
            "title": "Факторы риска (каждый = 1 балл)",
            "questions": [
                {
                    "id": "ihd",
                    "text": "ИБС в анамнезе",
                    "type": "boolean",
                    "description": "Инфаркт миокарда, положительный нагрузочный тест, использование нитратов, ЭКГ с патологическими Q-зубцами"
                },
                {
                    "id": "chf",
                    "text": "Сердечная недостаточность в анамнезе",
                    "type": "boolean",
                    "description": "Застойная сердечная недостаточность, отёк лёгких, пароксизмальная ночная одышка, ритм галопа S3"
                },
                {
                    "id": "cvd",
                    "text": "Цереброваскулярные заболевания",
                    "type": "boolean",
                    "description": "Инсульт или транзиторная ишемическая атака (ТИА) в анамнезе"
                },
                {
                    "id": "insulin_dm",
                    "text": "Сахарный диабет на инсулине",
                    "type": "boolean",
                    "description": "Диабет, требующий терапии инсулином до операции"
                },
                {
                    "id": "ckd",
                    "text": "Хроническая болезнь почек",
                    "type": "boolean",
                    "description": "Креатинин > 2 мг/дл (> 176.8 мкмоль/л)"
                },
                {
                    "id": "high_risk_surgery",
                    "text": "Операция высокого риска",
                    "type": "boolean",
                    "description": "Супраингвинальная сосудистая, интраперитонеальная или интраторакальная операция"
                }
            ]
        },
        {
            "section": "patient_info",
            "title": "Информация о пациенте",
            "questions": [
                {
                    "id": "age",
                    "text": "Возраст пациента (лет)",
                    "type": "number",
                    "min": 18,
                    "max": 120
                },
                {
                    "id": "creatinine",
                    "text": "Креатинин (мкмоль/л)",
                    "type": "number",
                    "min": 0,
                    "max": 2000,
                    "optional": true
                },
                {
                    "id": "surgery_description",
                    "text": "Описание планируемой операции",
                    "type": "text",
                    "optional": true
                }
            ]
        }
    ]$$::jsonb,
    $${
        "type": "sum",
        "fields": ["ihd", "chf", "cvd", "insulin_dm", "ckd", "high_risk_surgery"],
        "boolean_as_points": true
    }$$::jsonb,
    $${
        "ranges": [
            {"min": 0, "max": 0, "category": "Класс I (низкий риск)", "description": "0 факторов. Риск MACE: 3.9%"},
            {"min": 1, "max": 1, "category": "Класс II (промежуточный риск)", "description": "1 фактор. Риск MACE: 6.0%"},
            {"min": 2, "max": 2, "category": "Класс III (повышенный риск)", "description": "2 фактора. Риск MACE: 10.1%"},
            {"min": 3, "max": 6, "category": "Класс IV (высокий риск)", "description": "3+ факторов. Риск MACE: 15%+"}
        ],
        "note": "MACE = Major Adverse Cardiac Events (сердечная смерть, нефатальный ИМ, остановка сердца)"
    }$$::jsonb,
    1
)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    questions = EXCLUDED.questions,
    scoring_logic = EXCLUDED.scoring_logic,
    interpretation_rules = EXCLUDED.interpretation_rules,
    version = EXCLUDED.version,
    updated_at = NOW();

-- Goldman Cardiac Risk Index (original, for reference)
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000003',
    'GOLDMAN',
    'Goldman Cardiac Risk Index',
    'Оригинальный индекс кардиального риска Goldman для оценки периоперационных сердечных осложнений',
    'preoperative',
    $$[
        {
            "section": "history",
            "title": "Анамнез",
            "questions": [
                {"id": "age_over_70", "text": "Возраст > 70 лет", "type": "boolean", "score": 5},
                {"id": "mi_6mo", "text": "ИМ в последние 6 месяцев", "type": "boolean", "score": 10}
            ]
        },
        {
            "section": "physical",
            "title": "Физикальное обследование",
            "questions": [
                {"id": "s3_gallop", "text": "Ритм галопа S3 или расширение ярёмных вен", "type": "boolean", "score": 11},
                {"id": "aortic_stenosis", "text": "Значимый аортальный стеноз", "type": "boolean", "score": 3}
            ]
        },
        {
            "section": "ecg",
            "title": "ЭКГ",
            "questions": [
                {"id": "arrhythmia", "text": "Не синусовый ритм или ЖЭС на последней ЭКГ", "type": "boolean", "score": 7},
                {"id": "pvc", "text": "> 5 ЖЭС/мин в любое время до операции", "type": "boolean", "score": 7}
            ]
        },
        {
            "section": "general",
            "title": "Общее состояние",
            "questions": [
                {"id": "poor_general", "text": "PaO2 < 60, PaCO2 > 50, K < 3.0, HCO3 < 20, мочевина > 50, Cr > 3.0, патология печени", "type": "boolean", "score": 3}
            ]
        },
        {
            "section": "surgery",
            "title": "Операция",
            "questions": [
                {"id": "emergency", "text": "Экстренная операция", "type": "boolean", "score": 4},
                {"id": "major_surgery", "text": "Интраперитонеальная, интраторакальная или аортальная операция", "type": "boolean", "score": 3}
            ]
        }
    ]$$::jsonb,
    $${
        "type": "sum",
        "sections": ["history", "physical", "ecg", "general", "surgery"]
    }$$::jsonb,
    $${
        "ranges": [
            {"min": 0, "max": 5, "category": "Класс I", "description": "Риск серьёзных осложнений: 1%"},
            {"min": 6, "max": 12, "category": "Класс II", "description": "Риск серьёзных осложнений: 7%"},
            {"min": 13, "max": 25, "category": "Класс III", "description": "Риск серьёзных осложнений: 14%"},
            {"min": 26, "max": 100, "category": "Класс IV", "description": "Риск серьёзных осложнений: 78%"}
        ]
    }$$::jsonb,
    1
)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    questions = EXCLUDED.questions,
    scoring_logic = EXCLUDED.scoring_logic,
    interpretation_rules = EXCLUDED.interpretation_rules,
    version = EXCLUDED.version,
    updated_at = NOW();

-- Caprini Score for VTE risk
INSERT INTO survey_templates (id, code, name, description, category, questions, scoring_logic, interpretation_rules, version) VALUES
(
    '00000000-0000-0000-0000-000000000004',
    'CAPRINI',
    'Caprini Score (VTE Risk)',
    'Шкала Caprini для оценки риска венозных тромбоэмболических осложнений при хирургических вмешательствах',
    'preoperative',
    $$[
        {
            "section": "1_point",
            "title": "Факторы риска (1 балл каждый)",
            "questions": [
                {"id": "age_41_60", "text": "Возраст 41-60 лет", "type": "boolean", "score": 1},
                {"id": "minor_surgery", "text": "Малая операция", "type": "boolean", "score": 1},
                {"id": "bmi_over_25", "text": "ИМТ > 25 кг/м2", "type": "boolean", "score": 1},
                {"id": "edema", "text": "Отёки нижних конечностей", "type": "boolean", "score": 1},
                {"id": "varicose", "text": "Варикозные вены", "type": "boolean", "score": 1},
                {"id": "pregnancy", "text": "Беременность или послеродовый период", "type": "boolean", "score": 1},
                {"id": "miscarriage", "text": "Невынашивание беременности в анамнезе", "type": "boolean", "score": 1},
                {"id": "oc_hrt", "text": "Приём оральных контрацептивов или ЗГТ", "type": "boolean", "score": 1},
                {"id": "sepsis", "text": "Сепсис (< 1 мес)", "type": "boolean", "score": 1},
                {"id": "lung_disease", "text": "Тяжёлое заболевание лёгких, пневмония (< 1 мес)", "type": "boolean", "score": 1},
                {"id": "copd", "text": "ХОБЛ", "type": "boolean", "score": 1},
                {"id": "mi", "text": "ИМ", "type": "boolean", "score": 1},
                {"id": "chf_current", "text": "Застойная сердечная недостаточность (< 1 мес)", "type": "boolean", "score": 1},
                {"id": "bed_rest", "text": "Постельный режим в анамнезе", "type": "boolean", "score": 1},
                {"id": "ibd", "text": "Воспалительные заболевания кишечника", "type": "boolean", "score": 1}
            ]
        },
        {
            "section": "2_points",
            "title": "Факторы риска (2 балла каждый)",
            "questions": [
                {"id": "age_61_74", "text": "Возраст 61-74 лет", "type": "boolean", "score": 2},
                {"id": "major_surgery", "text": "Большая операция (> 45 мин)", "type": "boolean", "score": 2},
                {"id": "arthroscopy", "text": "Артроскопическая операция", "type": "boolean", "score": 2},
                {"id": "laparoscopy", "text": "Лапароскопическая операция (> 45 мин)", "type": "boolean", "score": 2},
                {"id": "malignancy", "text": "Злокачественное новообразование", "type": "boolean", "score": 2},
                {"id": "bed_rest_current", "text": "Постельный режим > 72 ч", "type": "boolean", "score": 2},
                {"id": "central_venous", "text": "Центральный венозный катетер", "type": "boolean", "score": 2}
            ]
        },
        {
            "section": "3_points",
            "title": "Факторы риска (3 балла каждый)",
            "questions": [
                {"id": "age_over_75", "text": "Возраст 75+ лет", "type": "boolean", "score": 3},
                {"id": "vte_history", "text": "ТГВ/ТЭЛА в анамнезе", "type": "boolean", "score": 3},
                {"id": "family_vte", "text": "Семейный анамнез ТГВ/ТЭЛА", "type": "boolean", "score": 3},
                {"id": "factor_v", "text": "Фактор V Лейден", "type": "boolean", "score": 3},
                {"id": "prothrombin", "text": "Мутация протромбина 20210A", "type": "boolean", "score": 3},
                {"id": "lupus", "text": "Волчаночный антикоагулянт", "type": "boolean", "score": 3},
                {"id": "anticardiolipin", "text": "Антикардиолипиновые антитела", "type": "boolean", "score": 3},
                {"id": "homocysteine", "text": "Повышенный гомоцистеин", "type": "boolean", "score": 3},
                {"id": "hit", "text": "ГИТ в анамнезе", "type": "boolean", "score": 3},
                {"id": "thrombophilia", "text": "Другая тромбофилия", "type": "boolean", "score": 3}
            ]
        },
        {
            "section": "5_points",
            "title": "Факторы риска (5 баллов каждый)",
            "questions": [
                {"id": "stroke", "text": "Инсульт (< 1 мес)", "type": "boolean", "score": 5},
                {"id": "arthroplasty", "text": "Эндопротезирование", "type": "boolean", "score": 5},
                {"id": "hip_fracture", "text": "Перелом бедра, таза или ноги", "type": "boolean", "score": 5},
                {"id": "spinal_injury", "text": "Травма спинного мозга (< 1 мес)", "type": "boolean", "score": 5}
            ]
        }
    ]$$::jsonb,
    $${
        "type": "sum",
        "sections": ["1_point", "2_points", "3_points", "5_points"]
    }$$::jsonb,
    $${
        "ranges": [
            {"min": 0, "max": 0, "category": "Очень низкий риск", "description": "Риск ВТЭ < 0.5%. Ранняя мобилизация."},
            {"min": 1, "max": 2, "category": "Низкий риск", "description": "Риск ВТЭ ~1.5%. Механическая профилактика."},
            {"min": 3, "max": 4, "category": "Умеренный риск", "description": "Риск ВТЭ ~3%. Фармакопрофилактика и/или механическая."},
            {"min": 5, "max": 100, "category": "Высокий риск", "description": "Риск ВТЭ ~6%. Фармакопрофилактика + механическая."}
        ]
    }$$::jsonb,
    1
)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    questions = EXCLUDED.questions,
    scoring_logic = EXCLUDED.scoring_logic,
    interpretation_rules = EXCLUDED.interpretation_rules,
    version = EXCLUDED.version,
    updated_at = NOW();
