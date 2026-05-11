export function formatCurrency(amount: number, currency = 'EUR'): string {
  if (currency === 'EUR') {
    const formatted = amount % 1 === 0 ? amount.toString() : amount.toFixed(2);
    return `€${formatted}`;
  }
  return new Intl.NumberFormat('uk-UA', { style: 'currency', currency }).format(amount);
}
