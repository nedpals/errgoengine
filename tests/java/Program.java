public class Program {
    public static void main(String[] args) {
        int[] numbers = {1, 2, 3, 4, 5};
        int index = 5; 

        int result = divide(numbers[index], 2);
        System.out.println("Result: " + result);
    }

    public static int divide(int a, int b) {
        return a / b;
    }
}