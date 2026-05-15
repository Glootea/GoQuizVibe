
var isDirty = false;
var confirmDeleteImage = "Delete image?";
var answerOptionsLabel = "Answer Options";
var correctAnswerLabel = "Correct Answer";
var addOptionText = "Add Option";
var optionPlaceholder = "Choice 1";

function submitQuizForm() {
  var form = document.getElementById("quizEditForm");
  if (!form) return;

  var formData = new FormData(form);
  fetch(form.action || form.getAttribute("hx-put"), {
    method: "PUT",
    body: formData,
    headers: {
      "HX-Request": "true",
    },
  })
    .then((response) => {
      if (response.ok || response.status === 200) {
        showSaveAlert();
        setTimeout(() => {
          location.reload();
        }, 1000);
      } else {
        showErrorAlert();
      }
    })
    .catch(() => {
      showErrorAlert();
    });
}

function handleQuestionEditSubmit(event) {
  event.preventDefault();
  var button = event.target;
  var quizID = button.getAttribute("data-quiz-id");
  var questionID = button.getAttribute("data-question-id");
  var container = button.closest('div[id^="editForm_"]');

  if (!container || !quizID || !questionID) {
    return;
  }

  var url = `/admin/quizzes/${quizID}/question/${questionID}`;
  var formData = new FormData();

  var inputs = container.querySelectorAll("input, select, textarea");
  inputs.forEach((input) => {
    if (input.name) {
      formData.append(input.name, input.value);
    }
  });

  fetch(url, {
    method: "PUT",
    body: formData,
    headers: {
      "HX-Request": "true",
    },
  })
    .then((response) => {
      if (response.ok || response.status === 200) {
        showSaveAlert();
        container.classList.add("hidden");
      } else {
        showErrorAlert();
      }
    })
    .catch(() => {
      showErrorAlert();
    });
}

function uploadQuestionImage(input) {
  var url = input.getAttribute("data-upload-url");
  var file = input.files[0];
  if (!file || !url) return;

  var formData = new FormData();
  formData.append("image", file);

  fetch(url, {
    method: "POST",
    body: formData,
    headers: {
      "HX-Request": "true",
    },
  })
    .then((response) => {
      if (response.ok || response.status === 200) {
        showSaveAlert();
        setTimeout(() => {
          location.reload();
        }, 1000);
      } else {
        showErrorAlert();
      }
    })
    .catch(() => {
      showErrorAlert();
    });
}

function deleteQuestionImage(button) {
  var url = button.getAttribute("data-delete-url");
  if (!url) return;
  if (!confirm(confirmDeleteImage)) return;

  fetch(url, {
    method: "DELETE",
    headers: {
      "HX-Request": "true",
    },
  })
    .then((response) => {
      var container = button.closest(".relative.group");
      if (response.ok || response.status === 200) {
        if (container) {
          container.remove();
        }
      } else {
        showErrorAlert();
      }
    })
    .catch(() => {
      showErrorAlert();
    });
}

function markDirty() {
  isDirty = true;
}

function setupDirtyTracking() {
  var form = document.getElementById("quizEditForm");
  if (!form) return;

  initialData = getFormData(form);

  var inputs = form.querySelectorAll("input, select, textarea");
  inputs.forEach((input) => {
    input.addEventListener("change", markDirty);
    input.addEventListener("input", markDirty);
  });
}

function getFormData(form) {
  var data = {};
  var inputs = form.querySelectorAll("input, select, textarea");
  inputs.forEach((input) => {
    if (input.name) {
      data[input.name] = input.value;
    }
  });
  return data;
}

window.addEventListener("beforeunload", (e) => {
  if (isDirty) {
    e.preventDefault();
    e.returnValue = "";
  }
});

function showSaveAlert() {
  var alert = document.getElementById("saveAlert");
  alert.classList.remove("hidden");
  setTimeout(() => {
    alert.classList.add("hidden");
  }, 2000);
}

function showErrorAlert() {
  var alert = document.getElementById("errorAlert");
  alert.classList.remove("hidden");
  setTimeout(() => {
    alert.classList.add("hidden");
  }, 2000);
}

function toggleEditForm(quizId, index) {
  var formId = `editForm_${quizId}_${index}`;
  var form = document.getElementById(formId);
  if (form) {
    form.classList.toggle("hidden");
  }
}

function updateAnswerOptions(select) {
  var container = select
    .closest(".space-y-3")
    .querySelector(".answer-options-container");
  if (!container) return;

  if (select.value === "choice") {
    container.innerHTML =
      '<label class="block mb-1 text-sm font-medium text-gray-700">' +
      answerOptionsLabel +
      "</label>" +
      '<div class="space-y-2 options-list">' +
      '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_0" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_0" placeholder="' +
      optionPlaceholder +
      '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
      '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_1" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_1" placeholder="' +
      optionPlaceholder +
      '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
      "</div>" +
      '<button type="button" onclick="addOption(this)" class="mt-2 text-sm text-indigo-600 hover:text-indigo-800"><i class="mr-1 fas fa-plus"></i>' +
      addOptionText +
      "</button>";
  } else if (select.value === "fill") {
    container.innerHTML =
      '<div class="fill-answer-editor">' +
      '<label class="block mb-2 text-sm font-medium text-gray-700">Правильные ответы (сегменты)</label>' +
      '<div id="segments-preview" class="p-3 bg-gray-50 rounded-lg text-sm mb-3">' +
      '<span class="text-gray-400">Добавьте текст и пропуски</span>' +
      '</div>' +
      '<div class="flex gap-2 mb-3">' +
      '<button type="button" onclick="addSegment(\'text\')" class="py-1 px-3 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg border"><i class="mr-1 fas fa-font"></i>Добавить текст</button>' +
      '<button type="button" onclick="addSegment(\'gap\')" class="py-1 px-3 text-sm bg-indigo-100 hover:bg-indigo-200 rounded-lg border border-indigo-300"><i class="mr-1 fas fa-question-circle"></i>Добавить пропуск</button>' +
      '</div>' +
      '<div id="segment-inputs" class="space-y-2"></div>' +
      '<input type="hidden" name="segments_json" id="segments-json"/>' +
      '</div>';
  } else {
    container.innerHTML =
      '<label class="block mb-1 text-sm font-medium text-gray-700">' +
      correctAnswerLabel +
      "</label>" +
      '<input type="text" name="correct_answer" class="py-2 px-3 w-full rounded-lg border focus:ring-2 focus:ring-indigo-500"/>';
  }
}

function updateAddFormAnswerOptions(select) {
  var container = select
    .closest("div[hx-post]")
    .querySelector("#addFormAnswerOptions");
  if (!container) return;

  if (select.value === "choice") {
    container.innerHTML =
      '<label class="block mb-1 text-sm font-medium text-gray-700">' +
      answerOptionsLabel +
      "</label>" +
      '<div id="addFormOptionsList" class="space-y-2">' +
      '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_0" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_0" placeholder="' +
      optionPlaceholder +
      '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
      '<div class="flex gap-2 items-center option-row"><input type="radio" name="correct_answer" value="option_1" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_1" placeholder="' +
      optionPlaceholder +
      '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button></div>' +
      "</div>" +
      '<button type="button" onclick="addOptionToForm()" class="mt-2 text-sm text-indigo-600 hover:text-indigo-800"><i class="mr-1 fas fa-plus"></i>' +
      addOptionText +
      "</button>";
  } else if (select.value === "fill") {
    container.innerHTML =
      '<div class="fill-answer-editor">' +
      '<label class="block mb-2 text-sm font-medium text-gray-700">Правильные ответы (сегменты)</label>' +
      '<div id="addForm-segments-preview" class="p-3 bg-gray-50 rounded-lg text-sm mb-3">' +
      '<span class="text-gray-400">Добавьте текст и пропуски</span>' +
      '</div>' +
      '<div class="flex gap-2 mb-3">' +
      '<button type="button" onclick="addSegment(\'text\')" class="py-1 px-3 text-sm bg-gray-100 hover:bg-gray-200 rounded-lg border"><i class="mr-1 fas fa-font"></i>Добавить текст</button>' +
      '<button type="button" onclick="addSegment(\'gap\')" class="py-1 px-3 text-sm bg-indigo-100 hover:bg-indigo-200 rounded-lg border border-indigo-300"><i class="mr-1 fas fa-question-circle"></i>Добавить пропуск</button>' +
      '</div>' +
      '<div id="addForm-segment-inputs" class="space-y-2"></div>' +
      '<input type="hidden" name="segments_json" id="addForm-segments-json"/>' +
      '</div>';
  } else {
    container.innerHTML =
      '<label class="block mb-1 text-sm font-medium text-gray-700">' +
      correctAnswerLabel +
      "</label>" +
      '<input type="text" name="correct_answer" class="py-2 px-3 w-full rounded-lg border focus:ring-2 focus:ring-indigo-500"/>';
  }
}

function addOption(button) {
  var container = button.closest(".answer-options-container");
  var optionsList = container.querySelector(".options-list");
  if (!optionsList) return;
  var count = optionsList.querySelectorAll(".option-row").length;
  var newRow = document.createElement("div");
  newRow.className = "flex gap-2 items-center option-row";
  newRow.innerHTML =
    '<input type="radio" name="correct_answer" value="option_' +
    count +
    '" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_' +
    count +
    '" placeholder="' +
    optionPlaceholder +
    '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
  optionsList.appendChild(newRow);
}

function addOptionToForm() {
  var optionsList = document.getElementById("addFormOptionsList");
  if (!optionsList) return;
  var count = optionsList.querySelectorAll(".option-row").length;
  var newRow = document.createElement("div");
  newRow.className = "flex gap-2 items-center option-row";
  newRow.innerHTML =
    '<input type="radio" name="correct_answer" value="option_' +
    count +
    '" class="w-4 h-4 text-indigo-600"/><input type="text" name="option_' +
    count +
    '" placeholder="' +
    optionPlaceholder +
    '" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/><button type="button" onclick="removeOption(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
  optionsList.appendChild(newRow);
}

function removeOption(btn) {
  var row = btn.closest(".option-row");
  if (row && row.parentElement.querySelectorAll(".option-row").length > 2) {
    row.remove();
  }
}

function syncSegmentsPreview(containerId) {
  var targetContainer = containerId
    ? document.getElementById(containerId)
    : document.getElementById("segment-inputs");
  var previewId = containerId ? containerId.replace("segment-inputs", "segments-preview") : "segments-preview";
  var preview = document.getElementById(previewId);
  if (!targetContainer || !preview) return;

  var html = "";
  targetContainer.querySelectorAll(".segment-row").forEach(function (row) {
    var type = row.querySelector('input[name$="_type"]');
    var content = row.querySelector('input[name$="_content"]');
    if (type && content) {
      if (type.value === "text") {
        html += content.value;
      } else {
        html += '<span class="mx-1 px-2 py-0.5 bg-indigo-100 border border-indigo-300 rounded text-indigo-700">' + content.value + '</span>';
      }
    }
  });
  preview.innerHTML = html || '<span class="text-gray-400">Добавьте текст и пропуски</span>';
}

function addSegment(type, containerId) {
  var targetContainer = containerId
    ? document.getElementById(containerId)
    : document.getElementById("segment-inputs");
  if (!targetContainer) return;
  var index = targetContainer.querySelectorAll(".segment-row").length;
  var row = document.createElement("div");
  row.className = "flex gap-2 items-center segment-row";
  row.dataset.type = type;

  if (type === "text") {
    row.innerHTML =
      '<input type="hidden" name="segment_' + index + '_type" value="text"/>' +
      '<span class="px-2 py-1 text-sm bg-gray-100 border border-gray-300 rounded text-gray-700">Текст</span>' +
      '<input type="text" name="segment_' + index + '_content" placeholder="Введите текст" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/>' +
      '<button type="button" onclick="removeSegment(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
  } else {
    row.innerHTML =
      '<input type="hidden" name="segment_' + index + '_type" value="gap"/>' +
      '<span class="px-2 py-1 text-sm bg-indigo-100 border border-indigo-300 rounded text-indigo-700">Пропуск</span>' +
      '<input type="text" name="segment_' + index + '_content" placeholder="Правильный ответ" class="flex-1 py-2 px-3 rounded-lg border focus:ring-2 focus:ring-indigo-500"/>' +
      '<button type="button" onclick="removeSegment(this)" class="text-red-500 hover:text-red-700"><i class="fas fa-times"></i></button>';
  }
  targetContainer.appendChild(row);
  syncSegmentsPreview(containerId);
}

function removeSegment(btn, containerId) {
  var row = btn.closest(".segment-row");
  if (row) {
    row.remove();
    reindexSegments(containerId);
    syncSegmentsPreview(containerId);
  }
}

function reindexSegments(containerId) {
  var targetContainer = containerId
    ? document.getElementById(containerId)
    : document.getElementById("segment-inputs");
  if (!targetContainer) return;
  var rows = targetContainer.querySelectorAll(".segment-row");
  rows.forEach(function (row, i) {
    row.querySelectorAll("input").forEach(function (input) {
      input.name = input.name.replace(/segment_\d+/, "segment_" + i);
    });
  });
}

function updateSegmentsJSON() {
  var container = document.getElementById("segment-inputs");
  var jsonInput = document.getElementById("segments-json");
  if (!container || !jsonInput) return;

  var segments = [];
  container.querySelectorAll(".segment-row").forEach(function (row) {
    var type = row.querySelector('input[name$="_type"]');
    var content = row.querySelector('input[name$="_content"]');
    if (type && content) {
      segments.push({ type: type.value, content: content.value });
    }
  });
  jsonInput.value = JSON.stringify(segments);
}

function setupFillAnswerEditor(correctAnswer) {
  var container = document.getElementById("segment-inputs");
  if (!container) return;

  if (correctAnswer) {
    try {
      var segments = JSON.parse(correctAnswer);
      segments.forEach(function (seg) {
        addSegment(seg.type, "segment-inputs");
        var rows = container.querySelectorAll(".segment-row");
        var lastRow = rows[rows.length - 1];
        if (lastRow) {
          var contentInput = lastRow.querySelector('input[name$="_content"]');
          if (contentInput) {
            contentInput.value = seg.content;
          }
        }
      });
      syncSegmentsPreview("segment-inputs");
    } catch (e) { }
  }
}

document.addEventListener("DOMContentLoaded", () => {
  setupDirtyTracking();
});

