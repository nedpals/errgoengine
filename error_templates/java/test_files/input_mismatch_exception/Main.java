import java.util.Scanner;

public class Main {
    public static void main(String[] args) {
        Scanner scanner = new Scanner(System.in);
        System.out.print("Enter an integer: ");
        int number = scanner.nextInt(); // Error: InputMismatchException for non-integer input
        System.out.println("You entered: " + number);
    }
}
