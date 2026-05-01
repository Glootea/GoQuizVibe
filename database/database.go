package database

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/goquizvibe/config"
	"github.com/goquizvibe/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(
	c config.Config,
) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=5433 sslmode=disable TimeZone=Asia/Shanghai", c.Database.User, c.Database.Password, c.Database.DBName)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying db: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Quiz{},
		&models.Question{},
		&models.QuizAttempt{},
		&models.UserAnswer{},
		&models.QuizSession{},
	)
}

func SeedData(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("Seeding initial data...")

	hash, _ := bcrypt.GenerateFromPassword([]byte("teacher123"), bcrypt.DefaultCost)
	teacherID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	teacher := &models.User{
		ID:           teacherID,
		Name:         "Учитель",
		Email:        "teacher@example.com",
		PasswordHash: string(hash),
		Role:         models.RoleTeacher,
		CreatedAt:    time.Now(),
	}
	if err := db.Create(teacher).Error; err != nil {
		return fmt.Errorf("failed to create teacher: %w", err)
	}

	hash, _ = bcrypt.GenerateFromPassword([]byte("student123"), bcrypt.DefaultCost)
	studentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	student := &models.User{
		ID:           studentID,
		Name:         "Студент",
		Email:        "student@example.com",
		PasswordHash: string(hash),
		Role:         models.RoleStudent,
		CreatedAt:    time.Now(),
	}
	if err := db.Create(student).Error; err != nil {
		return fmt.Errorf("failed to create student: %w", err)
	}

	quizzes := []struct {
		ID          uuid.UUID
		Title       string
		Description string
		Subject     string
		Grade       int
		TimeLimit   int
		Questions   []struct {
			Text          string
			Type          models.QuestionType
			Options       []string
			CorrectAnswer string
			Explanation   string
			Points        int
		}
	}{
		{
			ID:          uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Title:       "Математика 5 класс: Дроби и десятичные дроби",
			Description: "Тест на знание обыкновенных и десятичных дробей",
			Subject:     "Математика",
			Grade:       5,
			TimeLimit:   30,
			Questions: []struct {
				Text          string
				Type          models.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Чему равна сумма 1/4 + 1/4?", models.QuestionTypeChoice, []string{"1/2", "2/4", "1/8", "1"}, "1/2", "При сложении дробей с одинаковым знаменателем складываем числители: 1+1=2, получаем 2/4 = 1/2", 10},
				{"Переведите десятичную дробь 0.75 в обыкновенную дробь.", models.QuestionTypeOpen, []string{}, "3/4", "0.75 = 75/100 = 3/4 (сокращаем на 25)", 10},
				{"Вычислите: 2.5 + 1.3", models.QuestionTypeChoice, []string{"3.8", "3.7", "3.9", "4.0"}, "3.8", "Десятые доли: 0.5 + 0.3 = 0.8, целые: 2 + 1 = 3", 10},
				{"Заполните пропуск: 1/3 = ___/9", models.QuestionTypeFill, []string{}, "3", "Чтобы 1/3 привести к знаменателю 9, нужно и числитель, и знаменатель умножить на 3", 10},
				{"Чему равна разность 3/5 - 1/5?", models.QuestionTypeChoice, []string{"2/5", "1/5", "2/10", "4/5"}, "2/5", "При вычитании дробей с одинаковым знаменателем вычитаем числители: 3-1=2", 10},
				{"Переведите дробь 1/8 в десятичную дробь.", models.QuestionTypeOpen, []string{}, "0.125", "1/8 = 0.125, так как 1 ÷ 8 = 0.125", 15},
				{"Какое число больше: 0.6 или 3/5?", models.QuestionTypeChoice, []string{"Они равны", "0.6", "3/5", "Невозможно сравнить"}, "Они равны", "3/5 = 0.6, это одно и то же число", 15},
				{"Вычислите: 5 - 1.75", models.QuestionTypeOpen, []string{}, "3.25", "5.00 - 1.75 = 3.25", 10},
			},
		},
		{
			ID:          uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Title:       "Математика 6 класс: Пропорции и отрицательные числа",
			Description: "Тест на знание пропорций, координат и действий с отрицательными числами",
			Subject:     "Математика",
			Grade:       6,
			TimeLimit:   30,
			Questions: []struct {
				Text          string
				Type          models.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Решите уравнение: x + 5 = 2", models.QuestionTypeOpen, []string{}, "-3", "x = 2 - 5 = -3", 10},
				{"Чему равно произведение (-3) × (-4)?", models.QuestionTypeChoice, []string{"12", "-12", "7", "-7"}, "12", "При умножении двух отрицательных чисел результат положительный: (-3)×(-4)=12", 10},
				{"Какое число соответствует точке A на координатной прямой? (изображение точки A слева от 0 на расстоянии 3 делений)", models.QuestionTypeFill, []string{}, "-3", "Точка слева от нуля на расстоянии 3 единиц имеет координату -3", 10},
				{"Найдите x из пропорции: 2/x = 4/10", models.QuestionTypeChoice, []string{"5", "4", "2", "8"}, "5", "x = 2 × 10 / 4 = 20/4 = 5", 15},
				{"Чему равна сумма (-7) + 3?", models.QuestionTypeOpen, []string{}, "-4", "-7 + 3 = -4 (движемся вправо по числовой прямой на 3 единицы)", 10},
				{"Вычислите: |-5| + |3|", models.QuestionTypeChoice, []string{"8", "2", "-2", "-8"}, "8", "|-5| = 5, |3| = 3, значит 5 + 3 = 8", 10},
				{"Какой знак нужно поставить: -8 ___ -5 (чтобы получить верное утверждение)", models.QuestionTypeFill, []string{}, "<", "-8 меньше -5, так как -8 лежит левее на числовой прямой", 10},
				{"Решите: (-10) ÷ (-2)", models.QuestionTypeChoice, []string{"5", "-5", "20", "-20"}, "5", "При делении двух отрицательных чисел результат положительный: (-10)÷(-2)=5", 15},
			},
		},
		{
			ID:          uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			Title:       "Математика 7 класс: Одночлены и многочлены",
			Description: "Тест на знание одночленов, многочленов и формул сокращённого умножения",
			Subject:     "Математика",
			Grade:       7,
			TimeLimit:   30,
			Questions: []struct {
				Text          string
				Type          models.QuestionType
				Options       []string
				CorrectAnswer string
				Explanation   string
				Points        int
			}{
				{"Упростите: 3x² × 2x", models.QuestionTypeOpen, []string{}, "6x³", "3×2=6, x²×x=x³ (при умножении степени складываются)", 10},
				{"Раскройте скобки: (x + 3)²", models.QuestionTypeChoice, []string{"x² + 6x + 9", "x² + 9", "x² + 6x + 6", "x² + 3x + 9"}, "x² + 6x + 9", "По формуле (a+b)² = a² + 2ab + b²: x² + 2·x·3 + 3² = x² + 6x + 9", 15},
				{"Приведите подобные слагаемые: 5a + 3b - 2a + b", models.QuestionTypeOpen, []string{}, "3a + 4b", "5a - 2a = 3a; 3b + b = 4b", 10},
				{"Чему равна степень одночлена 7x³y²?", models.QuestionTypeFill, []string{}, "5", "Степень одночлена = сумма степеней всех переменных: 3 + 2 = 5", 10},
				{"Разложите на множители: x² - 16", models.QuestionTypeChoice, []string{"(x-4)(x+4)", "(x-8)(x+8)", "(x-4)²", "(x+4)²"}, "(x-4)(x+4)", "x² - 16 = x² - 4² = (x-4)(x+4) по формуле разности квадратов", 15},
				{"Выполните умножение: (x - 2)(x + 3)", models.QuestionTypeOpen, []string{}, "x² + x - 6", "(x-2)(x+3) = x·x + x·3 - 2·x - 2·3 = x² + 3x - 2x - 6 = x² + x - 6", 15},
				{"Какое выражение называется одночленом?", models.QuestionTypeChoice, []string{"Произведение чисел и переменных", "Сумма одночленов", "Выражение с делением на переменную", "Уравнение"}, "Произведение чисел и переменных", "Одночлен — это произведение чисел, переменных и их степеней", 10},
				{"Упростите: (2x)³", models.QuestionTypeFill, []string{}, "8x³", "(2x)³ = 2³ · x³ = 8x³", 15},
			},
		},
	}

	for _, q := range quizzes {
		quiz := &models.Quiz{
			ID:          q.ID,
			Title:       q.Title,
			Description: q.Description,
			Subject:     q.Subject,
			Grade:       q.Grade,
			Status:      models.QuizStatusAvailable,
			TimeLimit:   q.TimeLimit,
			CreatedBy:   teacherID,
			CreatedAt:   time.Now(),
		}
		if err := db.Create(quiz).Error; err != nil {
			return fmt.Errorf("failed to create quiz %s: %w", q.ID, err)
		}

		for i, qq := range q.Questions {
			optionsJSON, _ := json.Marshal(qq.Options)
			question := &models.Question{
				ID:            uuid.New(),
				QuizID:        q.ID,
				Text:          qq.Text,
				Type:          qq.Type,
				Options:       optionsJSON,
				CorrectAnswer: qq.CorrectAnswer,
				Explanation:   qq.Explanation,
				Points:        qq.Points,
				OrderIndex:    i,
			}
			if err := db.Create(question).Error; err != nil {
				return fmt.Errorf("failed to create question: %w", err)
			}
		}
	}

	log.Println("Seeding completed")
	return nil
}
