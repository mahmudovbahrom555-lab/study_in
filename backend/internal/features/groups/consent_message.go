package groups

import (
	"fmt"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// consentMessage — уведомление родителю о заявке ребёнка, на языке родителя.
// Тексты на узбекском проверить носителю.
func consentMessage(lang domain.Language, child, group, teacher string) (title, body string) {
	switch lang {
	case domain.LanguageUz:
		return "Ruxsatingiz kerak",
			fmt.Sprintf("%s %s ustozning «%s» guruhiga qo'shilmoqchi. Ilovada ruxsat bering yoki rad eting.", child, teacher, group)
	case domain.LanguageUzCyrl:
		return "Рухсатингиз керак",
			fmt.Sprintf("%s %s устознинг «%s» гуруҳига қўшилмоқчи. Иловада рухсат беринг ёки рад этинг.", child, teacher, group)
	case domain.LanguageEn:
		return "Your consent is needed",
			fmt.Sprintf("%s wants to join «%s» with tutor %s. Allow or decline in the app.", child, group, teacher)
	default:
		return "Нужно ваше согласие",
			fmt.Sprintf("%s хочет заниматься в группе «%s» у репетитора %s. Разрешите или откажите в приложении.", child, group, teacher)
	}
}
