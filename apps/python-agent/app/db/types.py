from typing import TypedDict


class CategoryRow(TypedDict):
    id: int
    name: str
    description: str | None


class ExpenseRow(TypedDict):
    id: int
    category_id: int
    category_name: str
    amount: float          # Decimal -> float at the boundary
    description: str | None
    month: int
    year: int


class MonthlyTotalRow(TypedDict):
    month: int
    year: int
    total_spent: float


class MonthlyAverageRow(TypedDict):
    month: int
    year: int
    average_amount: float
    expense_count: int


class YearMonthRow(TypedDict):
    year: int
    month: int