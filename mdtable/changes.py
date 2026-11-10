"""表のモデルを突き合わせ、利用者に出す変更の要約とセルの一覧を作る。"""

import difflib

from .model import Cell


class CellChange(object):
    """変更したセルの位置（0 起点）と、変更の種類。"""

    def __init__(self, row, col, old_text, new_text, kind):
        self.row = row
        self.col = col
        self.old_text = old_text
        self.new_text = new_text
        self.kind = kind


class Changes(object):
    def __init__(self):
        self.added_rows = 0
        self.deleted_rows = 0
        self.added_cols = 0
        self.deleted_cols = 0
        self.cells = []

    def summary(self):
        parts = []
        count = len({(cell.row, cell.col) for cell in self.cells})
        if count:
            parts.append("セル %d 個を変更" % count)
        for count, unit, action in (
                (self.added_rows, "行", "追加"), (self.deleted_rows, "行", "削除"),
                (self.added_cols, "列", "追加"), (self.deleted_cols, "列", "削除")):
            if count:
                parts.append("%d %sを%s" % (count, unit, action))
        return "（%s）" % "、".join(parts) if parts else ""

    def details(self, limit=20):
        cells = {}
        for cell in self.cells:
            descriptions = cells.setdefault((cell.row, cell.col), [])
            if cell.kind == "文字":
                descriptions.append("「%s」→「%s」" % (_short_text(cell.old_text), _short_text(cell.new_text)))
            else:
                descriptions.append("%sを変更" % cell.kind)
        lines = ["  %s: %s" % (_address(row, col), "、".join(descriptions))
                 for (row, col), descriptions in list(cells.items())[:limit]]
        if len(cells) > limit:
            lines.append("  ほか %d 個" % (len(cells) - limit))
        return lines


def _address(row, col):
    letters = ""
    number = col + 1
    while number:
        number, digit = divmod(number - 1, 26)
        letters = chr(ord("A") + digit) + letters
    return "%s%d" % (letters, row + 1)


def _short_text(text):
    text = text.replace("\r\n", "\n").replace("\r", "\n").replace("\n", "⏎")
    return text[:30] + "…" if len(text) > 30 else text


def _corresponding(old, new):
    """同じ並びを目印に対応を決め、置換の余りを追加・削除として数える。"""
    pairs = []
    added = deleted = 0
    matcher = difflib.SequenceMatcher(None, old, new, autojunk=False)
    for _tag, i1, i2, j1, j2 in matcher.get_opcodes():
        pairs.extend(zip(range(i1, i2), range(j1, j2)))
        added += max(0, (j2 - j1) - (i2 - i1))
        deleted += max(0, (i2 - i1) - (j2 - j1))
    return pairs, added, deleted


def compare_models(old, new):
    """列、次に行を文字で突き合わせ、対応したセルの変更を返す。"""
    result = Changes()
    old_anchors = old.anchors()
    new_anchors = new.anchors()

    def text_at(anchors, row, col):
        cell = anchors.get((row, col))
        return cell.text if cell is not None else ""

    if old.cols == new.cols:
        columns = list(zip(range(old.cols), range(new.cols)))
    else:
        old_columns = [tuple(text_at(old_anchors, row, col) for row in range(old.rows))
                       for col in range(old.cols)]
        new_columns = [tuple(text_at(new_anchors, row, col) for row in range(new.rows))
                       for col in range(new.cols)]
        columns, result.added_cols, result.deleted_cols = _corresponding(old_columns, new_columns)
    old_rows = [tuple(text_at(old_anchors, row, col) for col, _new_col in columns)
                for row in range(old.rows)]
    new_rows = [tuple(text_at(new_anchors, row, col) for _old_col, col in columns)
                for row in range(new.rows)]
    rows, result.added_rows, result.deleted_rows = _corresponding(old_rows, new_rows)
    for old_row, new_row in rows:
        for old_col, new_col in columns:
            previous = old_anchors.get((old_row, old_col))
            current = new_anchors.get((new_row, new_col))
            # 結合に隠れた位置は数えず、左上のセルで範囲の変更を比べる
            if ((previous is None and old.cell_at(old_row, old_col) is not None)
                    or (current is None and new.cell_at(new_row, new_col) is not None)):
                continue
            previous = previous if previous is not None else Cell(old_row, old_col)
            current = current if current is not None else Cell(new_row, new_col)
            for kind, before, after in (
                    ("文字", previous.text, current.text),
                    ("配置", previous.align, current.align),
                    ("見出し", previous.header, current.header),
                    ("結合", (previous.rowspan, previous.colspan), (current.rowspan, current.colspan))):
                if before != after:
                    result.cells.append(CellChange(new_row, new_col, previous.text, current.text, kind))
    result.cells.sort(key=lambda cell: (cell.row, cell.col))
    return result
