(function() {
    var isDirty = false;
    var initialData = {};
    var confirmDeleteQuestion = "Delete question?";
    var confirmDeleteImage = "Delete image?";
    var answerOptionsLabel = "Answer Options";
    var correctAnswerLabel = "Correct Answer";
    var addOptionText = "Add Option";
    var optionPlaceholder = "Choice 1";

    function submitQuizForm() {
        var form = document.getElementById('quizEditForm');
        if (!form) return;

        var formData = new FormData(form);
        fetch(form.action || form.getAttribute('hx-put'), {
            method: 'PUT',
            body: formData,
            headers: {
                'HX-Request': 'true'
            }
        }).then(function(response) {
            if (response.ok || response.status === 200) {
                showSaveAlert();
                setTimeout(function() {
                    location.reload();
                }, 1000);
            } else {
                showErrorAlert();
            }
        }).catch(function() {
            showErrorAlert();
        });
    }

    function handleQuestionEditSubmit(event) {
        event.preventDefault();
        var button = event.target;
        var quizID = button.getAttribute('data-quiz-id');
        var questionID = button.getAttribute('data-question-id');
        var container = button.closest('div[id^="editForm_"]');

        if (!container || !quizID || !questionID) {
            return;
        }

        var url = '/admin/quizzes/' + quizID + '/question/' + questionID;
        var formData = new FormData();

        var inputs = container.querySelectorAll('input, select, textarea');
        inputs.forEach(function(input) {
            if (input.name) {
                formData.append(input.name, input.value);
            }
        });

        fetch(url, {
            method: 'PUT',
            body: formData,
            headers: {
                'HX-Request': 'true'
            }
        }).then(function(response) {
            if (response.ok || response.status === 200) {
                showSaveAlert();
                container.classList.add('hidden');
            } else {
                showErrorAlert();
            }
        }).catch(function(error) {
            showErrorAlert();
        });
    }

    function uploadQuestionImage(input) {
        var url = input.getAttribute('data-upload-url');
        var file = input.files[0];
        if (!file || !url) return;

        var formData = new FormData();
        formData.append('image', file);

        fetch(url, {
            method: 'POST',
            body: formData,
            headers: {
                'HX-Request': 'true'
            }
        }).then(function(response) {
            if (response.ok || response.status === 200) {
                showSaveAlert();
                setTimeout(function() {
                    location.reload();
                }, 1000);
            } else {
                showErrorAlert();
            }
        }).catch(function() {
            showErrorAlert();
        });
    }

    function deleteQuestionImage(button) {
        var url = button.getAttribute('data-delete-url');
        if (!url) return;
        if (!confirm(confirmDeleteImage)) return;

        fetch(url, {
            method: 'DELETE',
            headers: {
                'HX-Request': 'true'
            }
        }).then(function(response) {
            if (response.ok || response.status === 200) {
                var container = button.closest('.relative.group');
                if (container) {
                    container.remove();
                }
            } else {
                showErrorAlert();
            }
        }).catch(function() {
            showErrorAlert();
        });
    }

    function markDirty() {
        isDirty = true;
    }

    function setupDirtyTracking() {
        var form = document.getElementById('quizEditForm');
        if (!form) return;

        initialData = getFormData(form);

        var inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach(function(input) {
            input.addEventListener('change', markDirty);
            input.addEventListener('input', markDirty);
        });
    }

    function getFormData(form) {
        var data = {};
        var inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach(function(input) {
            if (input.name) {
                data[input.name] = input.value;
            }
        });
        return data;
    }

    window.addEventListener('beforeunload', function(e) {
        if (isDirty) {
            e.preventDefault();
            e.returnValue = '';
        }
    });

    function showSaveAlert() {
        var alert = document.getElementById('saveAlert');
        alert.classList.remove('hidden');
        setTimeout(function() {
            alert.classList.add('hidden');
        }, 2000);
    }

    function showErrorAlert() {
        var alert = document.getElementById('errorAlert');
        alert.classList.remove('hidden');
        setTimeout(function() {
            alert.classList.add('hidden');
        }, 2000);
    }

    function toggleEditForm(quizId, index) {
        var formId = 'editForm_' + quizId + '_' + index;
        var form = document.getElementById(formId);
        if (form) {
            form.classList.toggle('hidden');
        }
    }

    function updateAnswerOptions(select) {
        var container = select.closest('.space-y-3').querySelector('.answer-options-container');
        if (!container) return;

        if (select.value === 'choice') {
            container.innerHTML = '<label class="block mb-1 text-sm font-medium text-gray-700">' + answerOptionsLabel + '</label>' +
                '<div class="space-y-2 options-list">' +
                '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_0" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_0" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
                '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_1" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_1" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
                '</div>' +
                '<button type="button" onclick="addOption(this)" class="mt-2 text-sm text-indigo-600 hover:text-indigo-800"><i class="mr-1 fas fa-plus"></i>' + addOptionText + '</button>';
        } else {
            container.innerHTML = '<label class="block mb-1 text-sm font-medium text-gray-700">' + correctAnswerLabel + '</label>' +
                '<input type="text" name="correct_answer" class="py-2 px-3 w-full rounded-lg border focus:ring-2 focus:ring-indigo-500"/>';
        }
    }

    function updateAddFormAnswerOptions(select) {
        var container = select.closest('div[hx-post]').querySelector('#addFormAnswerOptions');
        if (!container) return;

        if (select.value === 'choice') {
            container.innerHTML = '<label class="block mb-1 text-sm font-medium text-gray-700">' + answerOptionsLabel + '</label>' +
                '<div id="addFormOptionsList" class="space-y-2">' +
                '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_0" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_0" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
                '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_1" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_1" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
                '</div>' +
                '<button type="button" onclick="addOptionToForm()" class="mt-2 text-sm text-indigo-600 hover:text-indigo-800"><i class="mr-1 fas fa-plus"></i>' + addOptionText + '</button>';
        } else {
            container.innerHTML = '<label class="block mb-1 text-sm font-medium text-gray-700">' + correctAnswerLabel + '</label>' +
                '<input type="text" name="correct_answer" class="py-2 px-3 w-full rounded-lg border focus:ring-2 focus:ring-indigo-500"/>';
        }
    }

    function addOption(button) {
        var container = button.closest('.answer-options-container');
        var optionsList = container.querySelector('.options-list');
        if (!optionsList) return;
        var count = optionsList.querySelectorAll('.option-row').length;
        var newRow = document.createElement('div');
        newRow.className = 'flex gap-2 items-center option-row';
        newRow.innerHTML = '<input type="radio" name="correct_answer" value="option_' + count + '" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_' + count + '" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
        optionsList.appendChild(newRow);
    }

    function addOptionToForm() {
        var optionsList = document.getElementById('addFormOptionsList');
        if (!optionsList) return;
        var count = optionsList.querySelectorAll('.option-row').length;
        var newRow = document.createElement('div');
        newRow.className = 'flex gap-2 items-center option-row';
        newRow.innerHTML = '<input type="radio" name="correct_answer" value="option_' + count + '" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_' + count + '" placeholder="' + optionPlaceholder + '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
        optionsList.appendChild(newRow);
    }

    function removeOption(btn) {
        var row = btn.closest('.option-row');
        if (row && row.parentElement.querySelectorAll('.option-row').length > 2) {
            row.remove();
        }
    }

    document.addEventListener('DOMContentLoaded', function() {
        setupDirtyTracking();
    });
})();