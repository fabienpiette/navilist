'use strict';

function getFieldType(fieldValue) {
    var f = FIELDS.find(function(f) { return f.value === fieldValue; });
    return f ? f.type : 'text';
}

function buildSelect(name, options, selected) {
    var sel = document.createElement('select');
    sel.name = name;
    options.forEach(function(o) {
        var opt = document.createElement('option');
        opt.value = o.value;
        opt.textContent = o.label;
        if (o.value === selected) opt.selected = true;
        sel.appendChild(opt);
    });
    return sel;
}

function addRuleRow(field, op, value) {
    var fieldVal = field || 'title';
    var opVal = op || 'contains';
    var valStr = value !== undefined ? String(value) : '';

    var list = document.getElementById('rule-list');
    var row = document.createElement('div');
    row.className = 'rule-row';

    var fieldSel = buildSelect('rule_field', FIELDS.map(function(f) {
        return {value: f.value, label: f.label};
    }), fieldVal);
    fieldSel.addEventListener('change', function() {
        var newType = getFieldType(this.value);
        var opSel = row.querySelector('[name=rule_op]');
        while (opSel.firstChild) opSel.removeChild(opSel.firstChild);
        (OPERATORS[newType] || OPERATORS.text).forEach(function(o) {
            var opt = document.createElement('option');
            opt.value = o.value;
            opt.textContent = o.label;
            opSel.appendChild(opt);
        });
    });

    var opSel = buildSelect('rule_op', OPERATORS[getFieldType(fieldVal)] || OPERATORS.text, opVal);

    var valInput = document.createElement('input');
    valInput.type = 'text';
    valInput.name = 'rule_value';
    valInput.value = valStr;
    valInput.placeholder = 'value';

    var removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'action-btn danger remove-rule';
    removeBtn.textContent = '\u00d7';
    removeBtn.addEventListener('click', function() { row.remove(); });

    row.appendChild(fieldSel);
    row.appendChild(opSel);
    row.appendChild(valInput);
    row.appendChild(removeBtn);
    list.appendChild(row);
}

function visualToJSON() {
    var rows = document.querySelectorAll('.rule-row');
    var all = Array.prototype.map.call(rows, function(row) {
        var field = row.querySelector('[name=rule_field]').value;
        var op = row.querySelector('[name=rule_op]').value;
        var value = row.querySelector('[name=rule_value]').value;
        var criterion = {};
        criterion[op] = {};
        criterion[op][field] = value;
        return criterion;
    });
    var sort = document.getElementById('sort');
    var limit = document.getElementById('limit');
    return JSON.stringify({
        all: all,
        sort: sort ? sort.value : '+dateAdded',
        limit: limit ? parseInt(limit.value, 10) || 0 : 0
    }, null, 2);
}

function jsonToVisual() {
    try {
        var data = JSON.parse(document.getElementById('rules-json-area').value);
        var list = document.getElementById('rule-list');
        while (list.firstChild) list.removeChild(list.firstChild);
        (data.all || []).forEach(function(rule) {
            var opKey = Object.keys(rule)[0];
            var fieldObj = rule[opKey];
            var fieldKey = Object.keys(fieldObj)[0];
            addRuleRow(fieldKey, opKey, String(fieldObj[fieldKey]));
        });
        var sortEl = document.getElementById('sort');
        if (sortEl && data.sort) sortEl.value = data.sort;
        var limitEl = document.getElementById('limit');
        if (limitEl && data.limit !== undefined) limitEl.value = data.limit;
    } catch(e) {
        alert('Invalid JSON: ' + e.message);
    }
}

function switchEditorMode(mode) {
    var isJson = mode === 'json';
    document.getElementById('visual-editor').style.display = isJson ? 'none' : '';
    document.getElementById('json-editor').style.display = isJson ? '' : 'none';
    document.getElementById('editor-mode').value = mode;
    document.getElementById('btn-visual').classList.toggle('active', !isJson);
    document.getElementById('btn-json').classList.toggle('active', isJson);
    if (isJson) {
        document.getElementById('rules-json-area').value = visualToJSON();
    } else {
        jsonToVisual();
    }
}

document.addEventListener('DOMContentLoaded', function() {
    var data = typeof initialRulesJSON === 'string'
        ? JSON.parse(initialRulesJSON)
        : initialRulesJSON;
    (data.all || []).forEach(function(rule) {
        var opKey = Object.keys(rule)[0];
        var fieldObj = rule[opKey];
        var fieldKey = Object.keys(fieldObj)[0];
        addRuleRow(fieldKey, opKey, String(fieldObj[fieldKey]));
    });
    if (!data.all || data.all.length === 0) {
        addRuleRow('title', 'contains', '');
    }
    var ta = document.getElementById('rules-json-area');
    if (ta) ta.value = JSON.stringify(data, null, 2);
});
