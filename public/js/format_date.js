const pretty = str => {
    if (!str.includes('T')) return str;
    const [date, time] = str.split('T');
    return `${date} ${time.replace(/:\d\d(\.\d+)?(Z|[+-]\d\d:\d\d)?$/, '')}`;
  };