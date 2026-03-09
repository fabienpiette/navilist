'use strict';

// ─── Helpers ──────────────────────────────────────────────────────────────────

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

function buildFieldSelect(selected) {
    var sel = document.createElement('select');
    sel.name = 'rule_field';
    var groups = {}, groupOrder = [];
    FIELDS.forEach(function(f) {
        var g = f.group || '';
        if (!groups[g]) { groups[g] = []; groupOrder.push(g); }
        groups[g].push(f);
    });
    groupOrder.forEach(function(gname) {
        var container = sel;
        if (gname) {
            var og = document.createElement('optgroup');
            og.label = gname;
            sel.appendChild(og);
            container = og;
        }
        groups[gname].forEach(function(f) {
            var opt = document.createElement('option');
            opt.value = f.value;
            opt.textContent = f.label;
            if (f.value === selected) opt.selected = true;
            container.appendChild(opt);
        });
    });
    return sel;
}

function coerceValue(op, fieldType, rawValue) {
    if (op === 'inTheRange') {
        if (Array.isArray(rawValue)) return rawValue;
        var parts = String(rawValue).split(',');
        if (fieldType === 'number') return [Number(parts[0]) || 0, Number(parts[1]) || 0];
        return [(parts[0] || '').trim(), (parts[1] || '').trim()];
    }
    if (op === 'inTheLast' || op === 'notInTheLast') return Number(rawValue) || 0;
    if (fieldType === 'boolean') return rawValue === 'true' || rawValue === true;
    if (fieldType === 'number' && op !== 'before' && op !== 'after') { var n = Number(rawValue); return isNaN(n) ? rawValue : n; }
    return rawValue;
}

function displayValue(rawValue) {
    if (Array.isArray(rawValue)) return rawValue.join(',');
    return rawValue !== undefined ? String(rawValue) : '';
}

// ─── Autocomplete ─────────────────────────────────────────────────────────────

var AC_FIELDS = ['genre', 'artist', 'albumartist', 'album'];
var AC_OPS    = ['is', 'isNot', 'contains', 'notContains', 'startsWith', 'endsWith'];

function attachAutocomplete(wrap, fieldSel, opSel) {
    var input    = wrap.querySelector('input[name=rule_value]');
    var dropdown = document.createElement('div');
    dropdown.className = 'ac-dropdown';
    wrap.appendChild(dropdown);

    var timer     = null;
    var activeIdx = -1;

    function clearDropdown() {
        while (dropdown.firstChild) dropdown.removeChild(dropdown.firstChild);
    }

    function hideDropdown() {
        dropdown.style.display = 'none';
        clearDropdown();
        activeIdx = -1;
    }

    function showSuggestions(items) {
        clearDropdown();
        if (!items || !items.length) { hideDropdown(); return; }
        items.forEach(function(text) {
            var item = document.createElement('div');
            item.className = 'ac-item';
            item.textContent = text;
            item.addEventListener('mousedown', function(e) {
                e.preventDefault(); // keep focus on input so blur fires after selection
                input.value = text;
                hideDropdown();
            });
            dropdown.appendChild(item);
        });
        dropdown.style.display = 'block';
        activeIdx = -1;
    }

    function fetchSuggestions(q) {
        if (AC_FIELDS.indexOf(fieldSel.value) === -1) { hideDropdown(); return; }
        if (AC_OPS.indexOf(opSel.value) === -1)       { hideDropdown(); return; }
        if (!q) { hideDropdown(); return; }
        fetch('/smart/suggest?field=' + encodeURIComponent(fieldSel.value) + '&q=' + encodeURIComponent(q))
            .then(function(r) { return r.json(); })
            .then(showSuggestions)
            .catch(hideDropdown);
    }

    input.addEventListener('input', function() {
        clearTimeout(timer);
        timer = setTimeout(function() { fetchSuggestions(input.value); }, 220);
    });

    input.addEventListener('keydown', function(e) {
        var items = dropdown.querySelectorAll('.ac-item');
        if (e.key === 'Escape')    { hideDropdown(); return; }
        if (!items.length || dropdown.style.display === 'none') return;
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            activeIdx = Math.min(activeIdx + 1, items.length - 1);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            activeIdx = Math.max(activeIdx - 1, 0);
        } else if (e.key === 'Enter' && activeIdx >= 0) {
            e.preventDefault();
            input.value = items[activeIdx].textContent;
            hideDropdown();
            return;
        } else { return; }
        items.forEach(function(it, i) { it.classList.toggle('ac-active', i === activeIdx); });
    });

    input.addEventListener('blur', function() {
        setTimeout(hideDropdown, 150);
    });
}

// ─── Rule row ─────────────────────────────────────────────────────────────────

function buildRuleRow(field, op, value) {
    var fieldVal = field || 'title';
    var opVal = op || 'contains';

    var row = document.createElement('div');
    row.className = 'rule-row';

    var fieldSel = buildFieldSelect(fieldVal);

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
    valInput.value = displayValue(value);
    var fType = getFieldType(fieldVal);
    if (opVal === 'inTheRange') {
        valInput.placeholder = fType === 'date' ? 'YYYY-MM-DD,YYYY-MM-DD' : 'min,max';
    } else if (opVal === 'inTheLast' || opVal === 'notInTheLast') {
        valInput.placeholder = 'days';
    } else if (fType === 'date' || opVal === 'before' || opVal === 'after') {
        valInput.placeholder = 'YYYY-MM-DD';
    } else {
        valInput.placeholder = 'value';
    }

    var removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'action-btn danger';
    removeBtn.textContent = '\u00d7';
    removeBtn.title = 'Remove this rule';
    removeBtn.addEventListener('click', function() { row.remove(); });

    var acWrap = document.createElement('div');
    acWrap.className = 'ac-wrap';
    acWrap.appendChild(valInput);

    row.appendChild(fieldSel);
    row.appendChild(opSel);
    row.appendChild(acWrap);
    row.appendChild(removeBtn);

    attachAutocomplete(acWrap, fieldSel, opSel);
    return row;
}

function addRuleRow(field, op, value, container) {
    (container || document.getElementById('rule-list')).appendChild(
        buildRuleRow(field, op, value)
    );
}

// ─── Group block ──────────────────────────────────────────────────────────────

function buildGroup(type, rules) {
    var group = document.createElement('div');
    group.className = 'rule-group';

    // Header row
    var header = document.createElement('div');
    header.className = 'rule-group-header';

    var matchSpan = document.createElement('span');
    matchSpan.textContent = 'Match\u00a0';

    var typeSel = document.createElement('select');
    typeSel.className = 'group-type-select';
    [{value:'any',label:'any'},{value:'all',label:'all'}].forEach(function(o) {
        var opt = document.createElement('option');
        opt.value = o.value;
        opt.textContent = o.label;
        if (o.value === (type || 'any')) opt.selected = true;
        typeSel.appendChild(opt);
    });

    var ofSpan = document.createElement('span');
    ofSpan.textContent = '\u00a0of these rules:';

    var removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'action-btn danger';
    removeBtn.style.marginLeft = 'auto';
    removeBtn.textContent = '\u00d7';
    removeBtn.title = 'Remove this group';
    removeBtn.addEventListener('click', function() { group.remove(); });

    header.appendChild(matchSpan);
    header.appendChild(typeSel);
    header.appendChild(ofSpan);
    header.appendChild(removeBtn);

    // Sub-rules container
    var rulesContainer = document.createElement('div');
    rulesContainer.className = 'rule-group-rules';

    (rules || []).forEach(function(rule) {
        var opKey = Object.keys(rule)[0];
        var fieldObj = rule[opKey];
        if (Array.isArray(fieldObj)) return; // skip 2+ levels deep (not representable)
        var fieldKey = Object.keys(fieldObj)[0];
        rulesContainer.appendChild(buildRuleRow(fieldKey, opKey, fieldObj[fieldKey]));
    });

    if (!rules || rules.length === 0) {
        rulesContainer.appendChild(buildRuleRow('title', 'contains', ''));
    }

    // Add-rule button inside group
    var addBtn = document.createElement('button');
    addBtn.type = 'button';
    addBtn.className = 'tool-btn';
    addBtn.style.marginTop = '.5rem';
    addBtn.textContent = '+ add rule';
    addBtn.title = 'Add a rule to this group';
    addBtn.addEventListener('click', function() {
        rulesContainer.appendChild(buildRuleRow('title', 'contains', ''));
    });

    group.appendChild(header);
    group.appendChild(rulesContainer);
    group.appendChild(addBtn);
    return group;
}

function addGroup(type, rules) {
    document.getElementById('rule-list').appendChild(buildGroup(type, rules));
}

// ─── Collect criteria from a container (direct .rule-row children only) ───────

function buildCriterionFromRow(row) {
    var field = row.querySelector('[name=rule_field]').value;
    var op    = row.querySelector('[name=rule_op]').value;
    var raw   = row.querySelector('[name=rule_value]').value;
    var value = coerceValue(op, getFieldType(field), raw);
    var c = {}; c[op] = {}; c[op][field] = value;
    return c;
}

function collectFromContainer(container) {
    var result = [];
    var children = container.children;
    for (var i = 0; i < children.length; i++) {
        if (children[i].classList.contains('rule-row')) {
            result.push(buildCriterionFromRow(children[i]));
        }
    }
    return result;
}

// ─── Visual ↔ JSON ────────────────────────────────────────────────────────────

function visualToJSON() {
    var all = [];
    var children = document.getElementById('rule-list').children;
    for (var i = 0; i < children.length; i++) {
        var child = children[i];
        if (child.classList.contains('rule-group')) {
            var gtype = child.querySelector('.group-type-select').value;
            var subRules = collectFromContainer(child.querySelector('.rule-group-rules'));
            var g = {}; g[gtype] = subRules;
            all.push(g);
        } else if (child.classList.contains('rule-row')) {
            all.push(buildCriterionFromRow(child));
        }
    }
    var sortEl  = document.getElementById('sort');
    var orderEl = document.getElementById('order');
    var limitEl = document.getElementById('limit');
    return JSON.stringify({
        all:   all,
        sort:  sortEl  ? sortEl.value                      : 'dateadded',
        order: orderEl ? orderEl.value                     : 'asc',
        limit: limitEl ? parseInt(limitEl.value, 10) || 0 : 0
    }, null, 2);
}

function loadRulesIntoList(rules, list) {
    while (list.firstChild) list.removeChild(list.firstChild);
    (rules || []).forEach(function(rule) {
        var opKey    = Object.keys(rule)[0];
        var fieldObj = rule[opKey];
        if (Array.isArray(fieldObj)) {
            list.appendChild(buildGroup(opKey, fieldObj));
        } else {
            var fieldKey = Object.keys(fieldObj)[0];
            list.appendChild(buildRuleRow(fieldKey, opKey, fieldObj[fieldKey]));
        }
    });
}

function jsonToVisual() {
    try {
        var data = JSON.parse(document.getElementById('rules-json-area').value);
        var list = document.getElementById('rule-list');
        loadRulesIntoList(data.all, list);
        var sortEl  = document.getElementById('sort');
        if (sortEl  && data.sort)            sortEl.value  = data.sort;
        var orderEl = document.getElementById('order');
        if (orderEl && data.order)           orderEl.value = data.order;
        var limitEl = document.getElementById('limit');
        if (limitEl && data.limit !== undefined) limitEl.value = data.limit;
    } catch(e) {
        alert('Invalid JSON: ' + e.message);
    }
}

function switchEditorMode(mode) {
    var isJson = mode === 'json';
    document.getElementById('visual-editor').style.display = isJson ? 'none' : '';
    document.getElementById('json-editor').style.display   = isJson ? ''     : 'none';
    document.getElementById('editor-mode').value = mode;
    document.getElementById('btn-visual').classList.toggle('active', !isJson);
    document.getElementById('btn-json').classList.toggle('active',    isJson);
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

    var list = document.getElementById('rule-list');
    loadRulesIntoList(data.all, list);

    if (!data.all || data.all.length === 0) {
        list.appendChild(buildRuleRow('title', 'contains', ''));
    }

    var sortEl  = document.getElementById('sort');
    if (sortEl  && data.sort)  sortEl.value  = data.sort;
    var orderEl = document.getElementById('order');
    if (orderEl && data.order) orderEl.value = data.order;

    var ta = document.getElementById('rules-json-area');
    if (ta) ta.value = JSON.stringify(data, null, 2);
});
