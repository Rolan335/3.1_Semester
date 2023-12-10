import csv


data = {}
with open('data/grades.csv', mode='r', newline='') as csv_file:
    csv_reader = csv.DictReader(csv_file)
    for row in csv_reader:
        last_name = row["Last name"].strip()
        first_name = row["First name"].strip()
        ssn = row["SSN"].strip()
        tests = [float(row[f'Test{i}']) for i in range(1, 5)]
        valid_tests = [test for test in tests if test >= 0]
        if valid_tests:
            average_grade = sum(valid_tests) / len(valid_tests)
        else:
            average_grade = 0.0

        data[(last_name, first_name, ssn)] = average_grade

with open('data/newGrades.csv', mode='w', newline='') as csv_file:
    fieldHeader = ["Last name", "First name", "SSN", "Average Grade"]
    csv_writer = csv.DictWriter(csv_file, fieldnames = fieldHeader)
    csv_writer.writeheader()
    for (last_name, first_name, ssn), average_grade in data.items():
        csv_writer.writerow(
            {'Last name': last_name, 'First name': first_name, 'SSN': ssn, 'Average Grade': average_grade})
print("Новый файл создан")
